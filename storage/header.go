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

	// Vars endpoints
	SaveVars(workspace string, layout string, varmap map[string]interface{}) error
	GetVars(workspace, layout string) (*types.Vars, error)

	// Layout endpoints.
	GetLayout(workspace, layout string) (*types.LayoutRecord, error)
	SaveLayout(workspace string, layout *types.LayoutRecord) error
	SetLayoutStatus(workspace, layout, status string) error

	CreateJob(id, workspace, origin_url, origin_method string) (*types.Job, error)
	UpdateJob(j *types.Job) error
	GetJob(workspace, jobId string) (*types.Job, error)
}
