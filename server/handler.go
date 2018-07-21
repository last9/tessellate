package server

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/dispatcher"
	"github.com/tsocial/tessellate/storage/types"
)

// Save Workspace under workspaces/ .
func (s *Server) SaveWorkspace(ctx context.Context, in *SaveWorkspaceRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	// Tree for workspace ID.
	tree := types.MakeTree(in.Id)

	// Create a new types.Workspace instance to be returned.
	workspace := types.Workspace(in.Id)
	if err := s.store.Save(&workspace, tree); err != nil {
		return nil, err
	}

	vars := types.Vars{}

	if in.Vars != nil {
		// Create vars instance.
		if err := vars.Unmarshal(in.Vars); err != nil {
			return nil, err
		}
	}

	// Save the workspace and the vars.
	if err := s.store.Save(&vars, tree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}

// Get workspace for the mentioned Workspace ID.
func (s *Server) GetWorkspace(ctx context.Context, in *GetWorkspaceRequest) (*Workspace, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	// Make tree for workspace ID.
	tree := types.MakeTree(in.Id)
	workspace := types.Workspace(in.Id)

	// Get the workspace that should exist.
	if err := s.store.Get(&workspace, tree); err != nil {
		return nil, err
	}

	// Get versions of the workspace.
	versions, err := s.store.GetVersions(&workspace, tree)
	if err != nil {
		return nil, err
	}

	// Get the vars for that workspace ID.
	vars := types.Vars{}
	if err := s.store.Get(&vars, tree); err != nil {
		return nil, err
	}

	bytes, _ := vars.Marshal()

	// Return the workspace, with latest version and vars.
	w := Workspace{Name: string(workspace), Vars: bytes, Version: versions[1], Versions: versions}
	return &w, err
}

// Saves the layout under the mentioned workspace ID.
func (s *Server) SaveLayout(ctx context.Context, in *SaveLayoutRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	// Marshal vars for layout.
	vars := types.Vars{}
	if len(in.Vars) > 0 {
		if err := vars.Unmarshal(in.Vars); err != nil {
			return nil, err
		}
	}
	// Make tree for layout inside the workspace.
	lTree := types.MakeTree(in.WorkspaceId, in.Id)

	// Save vars in the layout tree.
	if err := s.store.Save(&vars, lTree); err != nil {
		return nil, err
	}

	// Make tree for workspace ID dir.
	tree := types.MakeTree(in.WorkspaceId)

	// Unmarshal layout plan as map.
	p := map[string]json.RawMessage{}
	if err := json.Unmarshal(in.Plan, &p); err != nil {
		return nil, err
	}

	// Create layout instance to be saved for given ID and plan.
	layout := types.Layout{Id: in.Id, Plan: p, Status: int32(Status_INACTIVE)}

	// Save the layout.
	if err := s.store.Save(&layout, tree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}

// GET layout for given layout ID.
func (s *Server) GetLayout(ctx context.Context, in *LayoutRequest) (*Layout, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	// Make workspace and layout trees.
	wTree := types.MakeTree(in.WorkspaceId)
	tree := types.MakeTree(in.WorkspaceId, in.Id)
	layout := types.Layout{Id: in.Id}

	// GET the layout from the workspace tree.
	if err := s.store.Get(&layout, wTree); err != nil {
		return nil, err
	}

	// GET the vars from the layout tree.
	vars := types.Vars{}
	if err := s.store.Get(&vars, tree); err != nil {
		return nil, err
	}

	// Marshal plan and vars.
	pBytes, _ := json.Marshal(layout.Plan)
	vBytes, _ := vars.Marshal()

	// Return the layout instance.
	lay := Layout{
		Workspaceid: in.WorkspaceId,
		Id:          layout.Id,
		Status:      Status(layout.Status),
		Plan:        pBytes,
		Vars:        vBytes,
	}

	return &lay, nil
}

// Operation layout for APPLY and DESTROY operations on the layout.
func (s *Server) opLayout(wID, lID string, op int32, vars []byte, dry bool) (*JobStatus, error) {
	lyt := types.Layout{}
	tree := types.MakeTree(wID)
	layoutTree := types.MakeTree(wID, lID)

	// GET versions of the layout.
	versions, err := s.store.GetVersions(&lyt, tree)
	if err != nil {
		return nil, err
	}

	v := types.Vars{}

	// todo check if vars are empty.
	if vars != nil {
		// Unmarshal in vars in v.
		if err := v.Unmarshal(vars); err != nil {
			return nil, err
		}

		// Make tree for layout.
		lTree := types.MakeTree(wID, lID)

		// Save the vars for apply op, in the layout tree.
		if err := s.store.Save(&v, lTree); err != nil {
			return nil, err
		}
	}

	// GET the version for vars.
	varsVersions, err := s.store.GetVersions(&v, layoutTree)
	if err != nil {
		return nil, err
	}

	// Return the job instance for layout with latest version of vars and layout.
	j := types.Job{
		LayoutId:      lID,
		LayoutVersion: versions[len(versions)-2],
		Status:        int32(JobState_PENDING),
		VarsVersion:   varsVersions[len(varsVersions)-2],
		Op:            op,
		Dry:           dry,
	}

	// Save this job in workspace tree.
	if err := s.store.Save(&j, tree); err != nil {
		return nil, err
	}

	job := &JobStatus{Id: j.Id, Status: JobState(j.Status)}
	return job, dispatcher.Get().Dispatch(j.Id, wID)
}

// Apply layout job.
func (s *Server) ApplyLayout(ctx context.Context, in *ApplyLayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.opLayout(in.WorkspaceId, in.Id, int32(Operation_APPLY), in.Vars, in.Dry)
}

// Destroy layout job.
func (s *Server) DestroyLayout(ctx context.Context, in *ApplyLayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.opLayout(in.WorkspaceId, in.Id, int32(Operation_DESTROY), in.Vars, in.Dry)
}

// Abort job.
func (s *Server) AbortJob(ctx context.Context, in *JobRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return nil, nil
}

// Start watch.
func (s *Server) StartWatch(ctx context.Context, in *StartWatchRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.saveWatch(in.WorkspaceId, in.Id, in.SuccessCallback, in.FailureCallback)
}

// Stop watch.
func (s *Server) StopWatch(ctx context.Context, in *StopWatchRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.saveWatch(in.WorkspaceId, in.Id, "", "")
}

// Saves the watch under layout tree.
func (s *Server) saveWatch(wID, lID, success, failure string) (*Ok, error) {
	tree := types.MakeTree(wID, lID)

	// Create a watch instance.
	watch := types.Watch{
		SuccessURL: success,
		FailureURL: failure,
	}

	// Save the watch in layout tree.
	if err := s.store.Save(&watch, tree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}
