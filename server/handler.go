package server

import (
	"context"
	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/storage/types"
	"encoding/json"
	"strings"
)

func (s *Server) SaveWorkspace(ctx context.Context, in *SaveWorkspaceRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	var values types.Vars

	for k, v := range in.Vars {
		values[k] = v
	}

	if err := s.store.SaveWorkspace(strings.ToLower(in.Id), &values); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}


func (s *Server) GetWorkspace(ctx context.Context, in *GetWorkspaceRequest) (*Workspace, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	versionRecord, err := s.store.GetWorkspace(in.Id)

	w := Workspace{}

	w.Version = versionRecord.Version
	w.Versions = versionRecord.Versions
	if w.Vars["Data"], err = json.Marshal(&versionRecord.Data); err != nil{
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	w.Name = in.Id

	return &w, err
}

func (s *Server) GetWorkspaceLayouts(
	ctx context.Context, in *GetWorkspaceLayoutsRequest,
) (*Layouts, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return &Layouts{}, nil
}

func (s *Server) SaveLayout(ctx context.Context, in *SaveLayoutRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	var values types.Vars

	for k, v := range in.Vars {
		values[k] = v
	}

	plan := make(map[string]interface{}, 2)

	for k, v := range in.Plan {
		plan[k] = v
	}

	if err := s.store.SaveLayout(strings.ToLower(in.WorkspaceId), strings.ToLower(in.Id), plan, &values); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}

func (s *Server) GetLayout(ctx context.Context, in *LayoutRequest) (*Layout, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return nil, nil
}

func (s *Server) ApplyLayout(ctx context.Context, in *ApplyLayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return nil, nil
}

func (s *Server) DestroyLayout(ctx context.Context, in *LayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return nil, nil
}

func (s *Server) GetJob(ctx context.Context, in *JobRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return nil, nil
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
