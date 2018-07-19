package server

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/dispatcher"
	"github.com/tsocial/tessellate/storage/types"
)

func (s *Server) SaveWorkspace(ctx context.Context, in *SaveWorkspaceRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	tree := types.MakeTree(in.Id)

	workspace := types.Workspace(in.Id)
	if err := s.store.Save(&workspace, tree); err != nil {
		return nil, err
	}

	vars := types.Vars{}
	if err := json.Unmarshal(in.Vars, &vars); err != nil {
		return nil, err
	}

	if err := s.store.Save(&vars, tree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}

func (s *Server) GetWorkspace(ctx context.Context, in *GetWorkspaceRequest) (*Workspace, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	tree := types.MakeTree(in.Id)
	workspace := types.Workspace(in.Id)

	if err := s.store.Get(&workspace, tree); err != nil {
		return nil, err
	}

	versions, err := s.store.GetVersions(&workspace, tree)
	if err != nil {
		return nil, err
	}

	vars := types.Vars{}
	if err := s.store.Get(&vars, tree); err != nil {
		return nil, err
	}

	bytes, _ := vars.Marshal()
	w := Workspace{Name: string(workspace), Vars: bytes, Version: versions[1], Versions: versions}
	return &w, err
}

func (s *Server) SaveLayout(ctx context.Context, in *SaveLayoutRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	vars := types.Vars{}
	if err := json.Unmarshal(in.Vars, &vars); err != nil {
		return nil, err
	}

	plan := map[string]json.RawMessage{}

	tree := types.MakeTree(in.WorkspaceId)

	for k, v := range in.Plan {
		var value json.RawMessage
		if err := json.Unmarshal(v, &value); err != nil {
			return nil, err
		}

		plan[k] = value

	}

	layout := types.Layout{Id: in.Id, Plan: plan, Status: int32(Status_INACTIVE)}
	if err := s.store.Save(&layout, tree); err != nil {
		return nil, err
	}

	lTree := types.MakeTree(in.WorkspaceId, in.Id)
	if err := s.store.Save(&vars, lTree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}

func (s *Server) GetLayout(ctx context.Context, in *LayoutRequest) (*Layout, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	wTree := types.MakeTree(in.WorkspaceId)
	tree := types.MakeTree(in.WorkspaceId, in.Id)
	layout := types.Layout{Id: in.Id}

	if err := s.store.Get(&layout, wTree); err != nil {
		return nil, err
	}

	vars := types.Vars{}
	if err := s.store.Get(&vars, tree); err != nil {
		return nil, err
	}

	p := map[string][]byte{}
	var err error
	for k, v := range layout.Plan {
		p[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	b, err := vars.Marshal()
	if err != nil {
		return nil, err
	}

	lay := Layout{
		Workspaceid: in.WorkspaceId,
		Id:          layout.Id,
		Plan:        p,
		Vars:        b,
		Status:      Status(layout.Status),
	}

	return &lay, err
}

func (s *Server) opLayout(wID, lID string, op int32, vars []byte, dry bool) (*JobStatus, error) {
	lyt := types.Layout{}
	tree := types.MakeTree(wID)
	layoutTree := types.MakeTree(wID, lID)

	versions, err := s.store.GetVersions(&lyt, tree)
	if err != nil {
		return nil, err
	}

	v := types.Vars{}
	if s.store.Get(&v, layoutTree); err != nil {
		return nil, err
	}

	varsVersions, err := s.store.GetVersions(&v, layoutTree)
	if err != nil {
		return nil, err
	}

	j := types.Job{
		LayoutId:      lID,
		LayoutVersion: versions[len(versions)-2],
		Status:        int32(JobState_PENDING),
		VarsVersion:   varsVersions[len(varsVersions)-2],
		Op:            op,
		Dry:           dry,
	}

	if err := s.store.Save(&j, tree); err != nil {
		return nil, err
	}

	job := &JobStatus{Id: j.Id, Status: JobState(j.Status)}
	return job, dispatcher.Get().Dispatch(j.Id, wID)
}

func (s *Server) ApplyLayout(ctx context.Context, in *ApplyLayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.opLayout(in.WorkspaceId, in.Id, int32(Operation_APPLY), in.Vars, in.Dry)
}

func (s *Server) DestroyLayout(ctx context.Context, in *ApplyLayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.opLayout(in.WorkspaceId, in.Id, int32(Operation_DESTROY), in.Vars, in.Dry)
}

func (s *Server) AbortJob(ctx context.Context, in *JobRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return nil, nil
}

func (s *Server) StartWatch(ctx context.Context, in *StartWatchRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.saveWatch(in.WorkspaceId, in.Id, in.SuccessCallback, in.FailureCallback)
}

func (s *Server) StopWatch(ctx context.Context, in *StopWatchRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.saveWatch(in.WorkspaceId, in.Id, "", "")
}

func (s *Server) saveWatch(wid, id, success, failure string) (*Ok, error) {
	tree := types.MakeTree(wid, id)

	watch := types.Watch{
		SuccessURL: success,
		FailureURL: failure,
	}

	if err := s.store.Save(&watch, tree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}
