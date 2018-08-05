package server

import (
	"context"
	"encoding/json"

	"fmt"

	"log"
	"regexp"

	"path/filepath"

	"strings"

	"github.com/meson10/highbrow"
	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/dispatcher"
	"github.com/tsocial/tessellate/storage/types"
)

const (
	retry = 5
	EXT   = ".tf.json"
)

// SaveWorkspace under workspaces/ .
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

	if in.Providers != nil {
		// Create vars instance.
		if err := vars.Unmarshal(in.Providers); err != nil {
			return nil, err
		}
	}

	// Save the workspace and the vars.
	if err := s.store.Save(&vars, tree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}

// GetWorkspace for the mentioned Workspace ID.
func (s *Server) GetWorkspace(ctx context.Context, in *GetWorkspaceRequest) (*Workspace, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.getWorkspace(in.Id)
}

func (s *Server) GetAllWorkspaces(ctx context.Context, in *Ok) (*AllWorkspaces, error) {
	keys, err := s.store.GetKeys(types.WORKSPACE+"/", "/")
	if err != nil {
		return nil, err
	}

	var workspaces []*Workspace
	for _, k := range keys {
		splits := strings.Split(k, "/")
		if len(splits) != 3 {
			log.Printf("skipping %s, len = %d\n", k, len(splits))
			continue
		}

		w, err := s.getWorkspace(splits[1])
		if err != nil {
			log.Printf("error while fetching workspace: %s, %+v", splits[1], err)
			continue
		}

		workspaces = append(workspaces, w)
	}

	return &AllWorkspaces{Workspaces: workspaces}, nil
}

func checkExt(filename string) (bool, error) {
	var validExt = regexp.MustCompile(`.*` + EXT)
	if !validExt.MatchString(filename) {
		return false, errors.New("invalid extension")
	}
	return true, nil
}

func (s *Server) getWorkspace(id string) (*Workspace, error) {
	// Make tree for workspace ID.
	tree := types.MakeTree(id)
	workspace := types.Workspace(id)

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

	vars.RedactSecrets()
	bytes, _ := vars.Marshal()

	// Return the workspace, with latest version and vars.
	w := Workspace{Name: string(workspace), Vars: bytes, Version: versions[1], Versions: versions}
	return &w, err
}

// SaveLayout under the mentioned workspace ID.
func (s *Server) SaveLayout(ctx context.Context, in *SaveLayoutRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	// Make tree for workspace ID dir.
	tree := types.MakeTree(in.WorkspaceId)

	// Unmarshal layout plan as map.
	p := map[string]json.RawMessage{}
	if err := json.Unmarshal(in.Plan, &p); err != nil {
		return nil, err
	}

	// Check the extension of the file, and raise an error if the ext is not tf.json.
	var err error

	for k := range p {
		_, err = checkExt(k)
		if err != nil {
			return nil, err
		}
	}

	wVars := &types.Vars{}
	s.store.Get(wVars, tree)

	// Check if this workspace supports providers by default.
	// If a workspace already supplies provider, then you must supply an Alias.
	if err := providerConflict(p, wVars); err != nil {
		return nil, errors.Wrap(err, "Provider conflict")
	}

	// Create layout instance to be saved for given ID and plan.
	layout := types.Layout{Id: in.Id, Plan: p, Status: int32(Status_INACTIVE)}

	// Save the layout.
	if err := s.store.Save(&layout, tree); err != nil {
		return nil, err
	}

	return &Ok{}, nil
}

// GetLayout for given layout ID.
func (s *Server) GetLayout(ctx context.Context, in *LayoutRequest) (*Layout, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	// Make workspace and layout trees.
	wTree := types.MakeTree(in.WorkspaceId)
	layout := types.Layout{Id: in.Id}

	// GET the layout from the workspace tree.
	if err := s.store.Get(&layout, wTree); err != nil {
		return nil, err
	}

	// Marshal plan and vars.
	pBytes, _ := json.Marshal(layout.Plan)

	// Return the layout instance.
	lay := Layout{
		Workspaceid: in.WorkspaceId,
		Id:          layout.Id,
		Status:      Status(layout.Status),
		Plan:        pBytes,
	}

	return &lay, nil
}

// Operation layout for APPLY and DESTROY operations on the layout.
func (s *Server) opLayout(wID, lID string, op int32, vars []byte, dry bool) (*JobStatus, error) {
	lyt := types.Layout{Id: lID}
	tree := types.MakeTree(wID)
	layoutTree := types.MakeTree(wID, lID)

	// GET versions of the layout.
	versions, err := s.store.GetVersions(&lyt, tree)
	log.Print(versions)
	if err != nil {
		return nil, err
	}

	v := types.Vars{}

	var varID string
	// todo check if vars are empty.
	if vars != nil {
		// Unmarshal in vars in v.
		if err := v.Unmarshal(vars); err != nil {
			return nil, err
		}

		// Save the vars for apply op, in the layout tree.
		if err := s.store.Save(&v, layoutTree); err != nil {
			return nil, err
		}

		varID = map[string]interface{}(v)["id"].(string)
	}

	// Return the job instance for layout with latest version of vars and layout.
	j := types.Job{
		LayoutId:      lID,
		LayoutVersion: versions[len(versions)-2],
		Status:        int32(JobState_PENDING),
		VarsVersion:   varID,
		Op:            op,
		Dry:           dry,
	}

	// Save this job in workspace tree.
	if err := s.store.Save(&j, tree); err != nil {
		return nil, err
	}

	job := &JobStatus{Id: j.Id, Status: JobState(j.Status)}

	// Lock for workspace and layout.
	key := fmt.Sprintf("%v-%v", wID, lID)

	if err := highbrow.Try(retry, func() error {
		return s.store.Lock(key, job.Id)
	}); err != nil {
		return nil, err
	}

	return job, dispatcher.Get().Dispatch(j.Id, wID, j.LayoutId)
}

// ApplyLayout job.
func (s *Server) ApplyLayout(ctx context.Context, in *ApplyLayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.opLayout(in.WorkspaceId, in.Id, int32(Operation_APPLY), in.Vars, in.Dry)
}

// DestroyLayout job.
func (s *Server) DestroyLayout(ctx context.Context, in *DestroyLayoutRequest) (*JobStatus, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return s.opLayout(in.WorkspaceId, in.Id, int32(Operation_DESTROY), in.Vars, false)
}

// AbortJob to halt.
func (s *Server) AbortJob(ctx context.Context, in *JobRequest) (*Ok, error) {
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, Errors_INVALID_VALUE.String())
	}

	return nil, nil
}

// StartWatch to listen to state changes on a Layout
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

func (s *Server) GetState(ctx context.Context, in *GetStateRequest) (*GetStateResponse, error) {
	key := filepath.Join(types.STATE, in.WorkspaceId, in.LayoutId)
	data, err := s.store.GetKey(key)
	if err != nil {
		return nil, err
	}

	return &GetStateResponse{
		State: data,
	}, nil
}
