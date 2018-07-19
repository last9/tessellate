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

	values := types.Vars{}
	for k, v := range in.Vars {
		values[k] = v
	}

	if err := s.store.Save(&workspace, tree); err != nil {
		return nil, err
	}

	if err := s.store.Save(&values, tree); err != nil {
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

	vars := types.Vars{}

	versions, err := s.store.GetVersions(&vars, tree)
	if err != nil {
		return nil, err
	}

	if err := s.store.Get(&vars, tree); err != nil {
		return nil, err
	}

	mapVars := map[string][]byte{}
	for k, v := range vars {
		mapVars[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	w := Workspace{Name: string(workspace), Vars: mapVars, Version: versions[1], Versions: versions}

	return &w, err
}

func (s *Server) SaveLayout(ctx context.Context, in *SaveLayoutRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	values := types.Vars{}
	for k, v := range in.Vars {
		values[k] = v
	}

	plan := map[string]json.RawMessage{}

	tree := types.MakeTree(in.WorkspaceId, in.Id)

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

	if err := s.store.Save(&values, tree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}

func (s *Server) GetLayout(ctx context.Context, in *LayoutRequest) (*Layout, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	tree := types.MakeTree(in.WorkspaceId, in.Id)
	layout := types.Layout{Id: in.Id}

	if err := s.store.Get(&layout, tree); err != nil {
		return nil, err
	}

	vars := types.Vars{}
	if err := s.store.Get(&vars, tree); err != nil {
		return nil, err
	}

	var mapVars map[string][]byte

	var err error
	for k, v := range vars {
		mapVars[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	p := map[string][]byte{}
	for k, v := range layout.Plan {
		p[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	lay := Layout{Workspaceid: in.WorkspaceId, Id: layout.Id, Plan: p, Vars: mapVars, Status: Status(layout.Status)}

	return &lay, err
}

func (s *Server) ApplyLayout(ctx context.Context, in *ApplyLayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	values := types.Vars{}
	for k, v := range in.Vars {
		values[k] = v
	}

	lyt := types.Layout{}
	layoutTree := types.MakeTree(in.WorkspaceId, in.Id)
	tree := types.MakeTree(in.WorkspaceId)

	versions, err := s.store.GetVersions(&lyt, layoutTree)
	if err != nil {
		return nil, err
	}

	vars := types.Vars{}
	if s.store.Get(&vars, layoutTree); err != nil {
		return nil, err
	}

	varsVersions, err := s.store.GetVersions(&lyt, layoutTree)
	if err != nil {
		return nil, err
	}

	j := types.Job{LayoutId: lyt.Id, LayoutVersion: versions[len(versions)-2], Status: int32(JobState_PENDING), VarsVersion: varsVersions[len(varsVersions)-2],
		Op: int32(Operation_APPLY), Dry: false}

	if err := s.store.Save(&j, tree); err != nil {
		return nil, err
	}

	job := JobStatus{Id: j.Id, Status: JobState(j.Status)}

	return &job, dispatcher.Get().Dispatch(j.Id, in.WorkspaceId)
}

func (s *Server) DestroyLayout(ctx context.Context, in *LayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	lyt := types.Layout{}
	layoutTree := types.MakeTree(in.WorkspaceId, in.Id)
	tree := types.MakeTree(in.WorkspaceId)

	versions, err := s.store.GetVersions(&lyt, layoutTree)
	if err != nil {
		return nil, err
	}

	vars := types.Vars{}
	if s.store.Get(&vars, layoutTree); err != nil {
		return nil, err
	}

	varsVersions, err := s.store.GetVersions(&lyt, layoutTree)
	if err != nil {
		return nil, err
	}

	j := types.Job{LayoutId: lyt.Id, LayoutVersion: versions[1], Status: int32(JobState_PENDING),
		VarsVersion: varsVersions[1], Op: int32(Operation_DESTROY), Dry: false}

	if err := s.store.Save(&j, tree); err != nil {
		return nil, err
	}

	job := JobStatus{Id: j.Id, Status: JobState(j.Status)}

	return &job, dispatcher.Get().Dispatch(j.Id, in.WorkspaceId)
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

	tree := types.MakeTree(in.WorkspaceId, in.Id)
	layout := types.Layout{Id: in.Id}

	if err := s.store.Get(&layout, tree); err != nil {
		return nil, err
	}

	vars := types.Vars{}
	if err := s.store.Get(&vars, tree); err != nil {
		return nil, err
	}

	varsVersions, err := s.store.GetVersions(&layout, tree)
	if err != nil {
		return nil, err
	}

	watch := types.Watch{LayoutId: layout.Id, LayoutVersion: varsVersions[len(varsVersions)-2],
		SuccessURL: in.SuccessCallback, FailureURL: in.FailureCallback}

	if err := s.store.Save(&watch, tree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}

func (s *Server) StopWatch(ctx context.Context, in *StopWatchRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	tree := types.MakeTree(in.WorkspaceId, in.Id)
	watch := types.Watch{LayoutId: in.Id}

	if err := s.store.Get(&watch, tree); err != nil {
		return nil, err
	}

	watch = types.Watch{LayoutId: watch.Id, LayoutVersion: watch.LayoutVersion,
		SuccessURL: "", FailureURL: ""}

	if err := s.store.Save(&watch, tree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}
