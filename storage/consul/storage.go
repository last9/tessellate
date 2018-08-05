package consul

import (
	"log"
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

func (e *ConsulStore) GetKey(key string) ([]byte, error) {
	bytes, _, err := e.client.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if bytes == nil {
		return []byte{}, nil
	}

	return bytes.Value, err
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
	log.Println(path)
	// Get the vars for the layout.
	bytes, _, err := e.client.KV().Get(path, nil)
	if err != nil {
		return errors.Wrapf(err, "Cannot fetch object for %v", path)
	}

	if bytes == nil {
		return errors.Errorf("Missing Key %v", path)
	}

	if err := reader.Unmarshal(bytes.Value); err != nil {
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
	b, err := source.Marshal()
	if err != nil {
		return errors.Wrap(err, "Cannot Marshal vars")
	}

	ts := time.Now().UnixNano()
	key := source.MakePath(tree)

	latestKey := path.Join(key, "latest")
	timestampKey := path.Join(key, fmt.Sprintf("%+v", ts))

	session := types.MakeVersion()

	lock, err := e.client.LockKey(path.Join(key, "lock"))
	if err != nil {
		return errors.Wrap(err, "Cannot Lock key")
	}

	// Create a Tx Chain of Ops.
	ops := api.KVTxnOps{
		&api.KVTxnOp{
			Verb:    api.KVSet,
			Key:     latestKey,
			Value:   b,
			Session: session,
		},
		&api.KVTxnOp{
			Verb:    api.KVSet,
			Key:     timestampKey,
			Value:   b,
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

func (e *ConsulStore) GetClient() *api.Client {
	return e.client
}
