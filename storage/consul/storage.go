package consul

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"path"

	"time"

	"fmt"

	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/storage/types"
)

func MakeConsulStore(addr ...string) *ConsulStore {
	return &ConsulStore{addr: addr}
}

type ConsulStore struct {
	addr   []string
	client *api.Client
}

func (e *ConsulStore) get(key string) ([]byte, error) {
	b, _, err := e.client.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if b == nil {
		return []byte{}, nil
	}

	// check if response is valid json
	var res interface{}
	if err := json.Unmarshal(b.Value, &res); err == nil {
		return b.Value, err
	}

	// When the content is gzipped
	r, err := gzip.NewReader(bytes.NewReader(b.Value))
	if err != nil {
		return nil, fmt.Errorf("invalid gzip or json")
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return data, err
}

func (e *ConsulStore) gzip(unzipped []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(unzipped); err != nil {
		return nil, err
	}

	if err := gz.Flush(); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (e *ConsulStore) save(key string, value []byte) error {
	gz, err := e.gzip(value)
	if err != nil {
		return err
	}

	_, err = e.client.KV().Put(&api.KVPair{Key: key, Value: gz}, nil)
	return err
}

func (e *ConsulStore) GetKey(key string) ([]byte, error) {
	bytes, err := e.get(key)
	if err != nil {
		return nil, err
	}

	if bytes == nil {
		return []byte{}, nil
	}

	return bytes, err
}

func (e *ConsulStore) SaveKey(key string, value []byte) error {
	return e.save(key, value)
}

func (e *ConsulStore) GetVersions(reader types.ReaderWriter, tree *types.Tree) ([]string, error) {
	key := reader.MakePath(tree)
	l, _, err := e.client.KV().Keys(key, "", nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot list %v", key)
	}

	var keys []string
	for _, k := range l {
		splitByKey := strings.SplitAfter(k, key+"/")
		for _, k2 := range splitByKey {
			if !strings.Contains(k2, "/") {
				keys = append(keys, k2)
			}
		}
	}

	return keys, nil
}

func (e *ConsulStore) Get(reader types.ReaderWriter, tree *types.Tree) error {
	return e.GetVersion(reader, tree, "latest")
}

func (e *ConsulStore) GetKeys(prefix string, separator string) ([]string, error) {
	l, _, err := e.client.KV().Keys(prefix, separator, nil)
	return l, err
}

func (e *ConsulStore) GetVersion(reader types.ReaderWriter, tree *types.Tree, version string) error {
	path := path.Join(reader.MakePath(tree), version)
	// Get the vars for the layout.
	bytes, err := e.get(path)
	if err != nil {
		return errors.Wrapf(err, "Cannot fetch object for %v", path)
	}

	if bytes == nil || len(bytes) == 0 {
		return errors.Errorf("Missing Key %v", path)
	}

	if err := reader.Unmarshal(bytes); err != nil {
		return errors.Wrap(err, "Cannot unmarshal data into Reader")
	}

	return nil
}

// Internal method to save Any data under a hierarchy that follows revision control.
// Example: In a workspace staging, you wish to save a new layout called dc1
// saveRevision("staging", "layout", "dc1", {....}) will try to save the following structure
// workspace/layouts/dc1/latest
// workspace/layouts/dc1/new_timestamp
// NOTE: This is an atomic operation, so either everything is written or nothing is.
// The operation may take its own sweet time before a quorum write is guaranteed.
func (e *ConsulStore) Save(source types.ReaderWriter, tree *types.Tree) error {
	ts := time.Now().UnixNano()
	return e.SaveTag(source, tree, fmt.Sprintf("%+v", ts))
}

func (e *ConsulStore) SaveTag(source types.ReaderWriter, tree *types.Tree, ts string) error {
	b, err := source.Marshal()
	if err != nil {
		return errors.Wrap(err, "Cannot Marshal vars")
	}

	key := source.MakePath(tree)

	latestKey := path.Join(key, "latest")
	timestampKey := path.Join(key, ts)

	session := types.MakeVersion()

	lock, err := e.client.LockKey(path.Join(key, "lock"))
	if err != nil {
		return errors.Wrap(err, "Cannot Lock key")
	}

	// Create a Tx Chain of Ops.

	gz, err := e.gzip(b)
	if err != nil {
		return err
	}

	ops := api.KVTxnOps{
		&api.KVTxnOp{
			Verb:    api.KVSet,
			Key:     latestKey,
			Value:   gz,
			Session: session,
		},
		&api.KVTxnOp{
			Verb:    api.KVSet,
			Key:     timestampKey,
			Value:   gz,
			Session: session,
		},
	}

	ok, _, _, err := e.client.KV().Txn(ops, nil)

	if err != nil {
		return errors.Wrap(err, "Cannot save Consul Transaction")
	}

	if !ok {
		return errors.New("Txn was rolled back. Weird, huh!")
	}

	source.SaveId(fmt.Sprintf("%v", ts))

	lock.Unlock()
	return nil
}

func (e *ConsulStore) Setup() error {
	conf := api.DefaultConfig()
	if len(e.addr) > 0 {
		conf.Address = e.addr[0]
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return err
	}

	e.client = client
	return nil
}

func (e *ConsulStore) Lock(key, s string) error {
	ok, _, err := e.client.KV().CAS(&api.KVPair{
		Key:         path.Join("lock", key),
		ModifyIndex: 0,
		CreateIndex: 0,
		Value:       []byte(s),
	}, nil)

	if err != nil {
		return err
	}

	if !ok {
		return errors.New("Cannot write Lock")
	}

	return nil
}

func (e *ConsulStore) Unlock(key string) error {
	key = path.Join("lock", key)
	_, err := e.client.KV().Delete(key, nil)
	if err != nil {
		return err
	}

	return nil
}

func (e *ConsulStore) Teardown() error {
	return nil
}

func (e *ConsulStore) DeleteKeys(prefix string) error {
	_, err := e.client.KV().DeleteTree(prefix+"/", &api.WriteOptions{})
	return err
}
