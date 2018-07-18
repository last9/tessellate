package server

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/storage/types"
)

func (s *Server) SaveWorkspace(ctx context.Context, in *SaveWorkspaceRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	tree := types.MakeTree(in.Id)

	workspace := types.Workspace(in.Id)

	var values types.Vars

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

	var l map[string][]byte

	for k, v := range vars {
		l[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	w := Workspace{Name: string(workspace), Vars: l, Version: versions[1], Versions: versions}

	return &w, err
}

func (s *Server) SaveLayout(ctx context.Context, in *SaveLayoutRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}
	var values types.Vars

	for k, v := range in.Vars {
		values[k] = v
	}

	var plan map[string]interface{}

	tree := types.MakeTree(in.WorkspaceId, in.Id)

	var err error
	for k, v := range in.Plan {
		var value interface{}
		err = json.Unmarshal(v, &value)
		plan[k] = value

		if err != nil {
			return nil, err
		}
	}

	layout := types.Layout{Id: in.Id, Plan: plan, Status: Status_INACTIVE}

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

	var l map[string][]byte

	var err error
	for k, v := range vars {
		l[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	var p map[string][]byte

	for k, v := range layout.Plan {
		p[k], err = json.Marshal(v)
		if err != nil {
			return nil, err
		}
	}

	lay := Layout{Workspaceid: in.WorkspaceId, Id: layout.Id, Plan: p, Vars: l, Status: layout.Status}

	return &lay, err
}

func (s *Server) ApplyLayout(ctx context.Context, in *ApplyLayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}
	var values types.Vars

	// todo merge vars of in.Vars and the job Vars.

	for k, v := range in.Vars {
		values[k] = v
	}

	lyt := types.Layout{}
	layout_tree := types.MakeTree(in.WorkspaceId, in.Id)
	tree := types.MakeTree(in.WorkspaceId)

	versions, err := s.store.GetVersions(&lyt, layout_tree)
	if err != nil {
		return nil, err
	}

	vars := types.Vars{}
	if s.store.Get(&vars, layout_tree); err != nil {
		return nil, err
	}

	varsVersions, err := s.store.GetVersions(&lyt, layout_tree)
	if err != nil {
		return nil, err
	}

	j := types.Job{LayoutId: lyt.Id, LayoutVersion: versions[1], Status: JobState_PENDING, VarsVersion: varsVersions[1],
		Op: Operation_APPLY, Dry: false}

	if err := s.store.Save(&j, tree); err != nil {
		return nil, err
	}

	job := JobStatus{Id: j.Id, Status: j.Status}

	return &job, nil
}

func (s *Server) DestroyLayout(ctx context.Context, in *LayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	lyt := types.Layout{}
	layout_tree := types.MakeTree(in.WorkspaceId, in.Id)
	tree := types.MakeTree(in.WorkspaceId)

	versions, err := s.store.GetVersions(&lyt, layout_tree)
	if err != nil {
		return nil, err
	}

	vars := types.Vars{}
	if s.store.Get(&vars, layout_tree); err != nil {
		return nil, err
	}

	varsVersions, err := s.store.GetVersions(&lyt, layout_tree)
	if err != nil {
		return nil, err
	}

	j := types.Job{LayoutId: lyt.Id, LayoutVersion: versions[1], Status: JobState_PENDING, VarsVersion: varsVersions[1],
		Op: Operation_DESTROY, Dry: false}

	if err := s.store.Save(&j, tree); err != nil {
		return nil, err
	}

	job := JobStatus{Id: j.Id, Status: j.Status}

	return &job, nil
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

	return nil, nil
}

func (s *Server) StopWatch(ctx context.Context, in *StopWatchRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return nil, nil
}
