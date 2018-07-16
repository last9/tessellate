package consul

import (
	"path"

	"encoding/json"

	"time"

	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/tsocial/tessellate/storage/types"
	"github.com/pkg/errors"
	"os"
	"github.com/tsocial/tessellate/server"
)

const layout_dir = "layouts"

func MakeConsulStore(addr ...string) *ConsulStore {
	return &ConsulStore{addr: addr}
}

type ConsulStore struct {
	addr   []string
	client *api.Client
}

func (e *ConsulStore) getRevisions(workspace, dir, name string) ([]string, error) {
	key := path.Join(workspace, dir, name)
	l, _, err := e.client.KV().List(key, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot list %v", key)
	}

	v := []string{}
	for _, n := range l {
		v = append(v, string(n.Key))
	}

	return v, nil
}

// Internal method to save Any data under a hierarchy that follows revision control.
// Example: In a workspace staging, you wish to save a new layout called dc1
// saveRevision("staging", "layout", "dc1", {....}) will try to save the following structure
// workspace/layouts/dc1/latest
// workspace/layouts/dc1/new_timestamp
// NOTE: This is an atomic operation, so either everything is written or nothing is.
// The operation may take its own sweet time before a quorum write is guaranteed.
func (e *ConsulStore) saveRevision(workspace, dir, name string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Cannot Marshal vars")
	}

	ts := time.Now().UnixNano()
	key := path.Join(workspace, dir, name)

	latestKey := path.Join(key, "latest")
	timestampKey := path.Join(key, fmt.Sprintf("%+v", ts))

	session := types.MakeVersion()

	lock, err := e.client.LockKey(path.Join(key, "lock"))
	if err != nil {
		return errors.Wrap(err, "Cannot Lock key")
	}

	defer lock.Unlock()

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

	return nil
}

func (e *ConsulStore) SaveWorkspace(id string, vars *types.Vars) error {
	return e.saveRevision(id, "vars", "default", vars)
}

func (e *ConsulStore) GetWorkspace(id string) (*types.VersionRecord, error) {
	versions, err := e.getRevisions(id, "vars", "default")
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot fetch Revisions for %v", id)
	}

	key := path.Join(id, "vars", "default", "latest")
	kp, _, err := e.client.KV().Get(key, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot fetch latest vars for %v", id)
	}

	if kp == nil {
		return nil, errors.Errorf("Missing Key %v", key)
	}

	return &types.VersionRecord{
		Data:     kp.Value,
		Version:  "latest",
		Versions: versions,
	}, nil
}

func (e *ConsulStore) SaveLayout(workspace, name string, plan map[string]interface{}, vars *types.Vars) error {

	enc :=  json.NewEncoder(os.Stdout)
	err := enc.Encode(plan)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Save the vars for the layout.
	e.saveRevision(workspace + path.Join(layout_dir, name), "vars", "default", vars)
	// Save the status for the layout.
	e.saveRevision(workspace, path.Join(layout_dir , name), "status", server.Status_ACTIVE.String())
	// Save the layout plan.
	e.saveRevision(workspace, layout_dir, name, plan)

	return nil
}

func (e *ConsulStore) GetLayout(workspace, name string) (*types.LayoutRecord, error) {

	// workspace-name/layouts/layout-name
	id := path.Join(workspace, layout_dir, name)

	// Get the versions.
	versions, err := e.getRevisions(id, "vars", "default")
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot fetch Revisions for %v", )
	}

	vars_key := path.Join(id, "vars", "default", "latest")
	plan_key := path.Join(id, "status", "latest")

	// Get the vars for the layout.
	env, _, err := e.client.KV().Get(vars_key, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot fetch latest vars for %v", id)
	}

	if env == nil {
		return nil, errors.Errorf("Missing Key %v", vars_key)
	}

	// Get the latest plan for the layout.
	plan, _, err := e.client.KV().Get(plan_key, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot fetch latest plan for %v", id)
	}

	if plan == nil {
		return nil, errors.Errorf("Missing Key %v", plan_key)
	}

	return &types.LayoutRecord{
		Plan:     plan.Value,
		Status:   string(plan.Value),
		Version:  "latest",
		Versions: versions,
		Env:	  env.Value,
	}, nil
}

func (e *ConsulStore) GetLayoutStatus(workspace, name string) (string, error) {
	layout, err := e.GetLayout(workspace, name)
	if err != nil {
		return "", errors.Wrapf(err, "Cannot fetch status for Layout id: %v", name)
	}

	return string(layout.Status) , nil
}

func (e *ConsulStore) SetLayoutStatus(workspace, name, status string) error {

	enc :=  json.NewEncoder(os.Stdout)
	err := enc.Encode(name)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Save the status for the layout.
	// todo talina: What if the layout doesn't exist. Log errors.
	e.saveRevision(workspace, path.Join(layout_dir , name), "status", status)

	return nil
}

// Gets all versions of all layouts of the workspace.
func (e *ConsulStore) GetWorkspaceLayouts(workspace string) ([]map[string]interface{}, error) {

	names, err := e.getRevisions(workspace, layout_dir, "")
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot fetch Layouts for %v", )
	}

	var layout_revisions map[string]interface{}
	var layouts []map[string]interface{}
	var id string


	for _, name := range names {
		versions, err := e.getRevisions(workspace, layout_dir, name)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot fetch versions for layout %v", name)
		}

		for _, ver := range versions {
			id = path.Join(workspace, layout_dir, name, ver)

			plan, _, err := e.client.KV().Get(id, nil)
			if err != nil {
				return nil, errors.Wrapf(err, "Cannot fetch Layouts for %v of version %v",
					name, ver)
			}

			layout_revisions[name] = plan
		}

		layouts = append(layouts, layout_revisions)
	}

	return layouts, nil
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

func (e *ConsulStore) Teardown() error {
	return nil
}

func (e *ConsulStore) GetClient() *api.Client {
	return e.client
}