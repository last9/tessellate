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

	// Workspace endpoints.
	GetWorkspace(id string) (*types.VersionRecord, error)
	SaveWorkspace(id string, vars *types.Vars) error
	GetWorkspaceLayouts(workspace string) ([]map[string]interface{}, error)

	SaveVars(workspace string, layout string, varmap map[string]interface{}) error
	GetVars(workspace, layout string) (*types.Vars, error)

	// Layout endpoints.
	GetLayout(workspace, layout string) (*types.LayoutRecord, error)
	SaveLayout(workspace string, layout *types.LayoutRecord) error
	SetLayoutStatus(workspace, layout, status string) error

	//
	//CreateTfRun(workspace, layout, varmap, layout_plan_id string) (*TfRun, error)
	//UpdateTfRun(*TfRun) error
	//GetTfRun(workspace, layout, run string) (*TfRun, error)
	//
	//CreateJob(id, workspace, origin_url, origin_method string) (*Job, error)
	//UpdateJob(j *Job) error
	//GetJob(workspace, id string) (*Job, error)
	//
	//GetState(workspace, v string) (*TfState, error)
	//GetTfRunState(workspace, layout, run string) (*TfState, error)
	//SaveState(workspace, v, id string, s *TfState, serial int) error
	//
	//GetLastTfRun(workspace, v string) (*TfRun, error)
	//
	//AcquireLock(key string) error
	//IsLocked(key string) bool
	//ReleaseLock(key string) error
}
