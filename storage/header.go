package storage

import (
	"github.com/tsocial/tessellate/storage/types"
	"github.com/hashicorp/consul/api"
)

func IdToVer(val string) (string, error) {
	return val, nil
}

type Storer interface {
	Setup() error
	Teardown() error
	GetClient() *api.Client

	GetWorkspace(id string) (*types.VersionRecord, error)
	SaveWorkspace(id string, vars *types.Vars) error

	//GetAllVars(env string) ([]map[string]interface{}, error)
	//SaveVars(env string, name string, varmap map[string]interface{}) error
	//GetVars(env, v string) (*types.VersionRecord, error)
	//
	//GetAllLayouts(env string) ([]map[string]interface{}, error)
	//GetLayout(env, v string) (*types.VersionRecord, error)
	SaveLayout(workspace, name string, layout map[string]interface{}, vars *types.Vars) error
	//GetLayoutStatus(env, layout string) (string, error)
	//SetLayoutStatus(env, layout, status string) error
	//
	//CreateTfRun(env, layout, varmap, layout_plan_id string) (*TfRun, error)
	//UpdateTfRun(*TfRun) error
	//GetTfRun(env, layout, run string) (*TfRun, error)
	//
	//CreateJob(id, env, origin_url, origin_method string) (*Job, error)
	//UpdateJob(j *Job) error
	//GetJob(env, id string) (*Job, error)
	//
	//GetState(env, v string) (*TfState, error)
	//GetTfRunState(env, layout, run string) (*TfState, error)
	//SaveState(env, v, id string, s *TfState, serial int) error
	//
	//GetLastTfRun(env, v string) (*TfRun, error)
	//
	//AcquireLock(key string) error
	//IsLocked(key string) bool
	//ReleaseLock(key string) error
}
