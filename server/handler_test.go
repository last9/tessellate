package server

import (
	"encoding/json"
	"testing"

	"context"
	"fmt"
	"io/ioutil"

	"github.com/tsocial/tessellate/storage"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/dispatcher"
	"github.com/tsocial/tessellate/storage/types"
	"github.com/tsocial/tessellate/utils"
)

var store storage.Storer
var server TessellateServer

func uglyJson(b []byte) []byte {
	var t map[string]interface{}
	json.Unmarshal(b, &t)
	b2, _ := json.Marshal(t)
	return b2
}

func TestServer_SaveAndGetWorkspace(t *testing.T) {
	id := fmt.Sprintf("workspace-%s", utils.RandString(8))

	t.Run("Should save a workspace.", func(t *testing.T) {
		req := &SaveWorkspaceRequest{Id: id}
		if _, err := server.SaveWorkspace(context.Background(), req); err != nil {
			errors.Wrap(err, Errors_INVALID_VALUE.String())
		}
	})

	t.Run("Should get the same workspace that was created.", func(t *testing.T) {
		req := &GetWorkspaceRequest{Id: id}
		resp, err := server.GetWorkspace(context.Background(), req)

		if err != nil {
			errors.Wrap(err, Errors_INVALID_VALUE.String())
		}
		fmt.Print(resp.String())
	})
}

func TestServer_SaveAndGetLayout(t *testing.T) {
	workspaceId := fmt.Sprintf("workspace-%s", utils.RandString(8))
	layoutId := fmt.Sprintf("layout-%s", utils.RandString(8))

	jobQueue := dispatcher.NewInMemory()
	dispatcher.Set(jobQueue)

	plan := map[string]json.RawMessage{}

	lBytes, err := ioutil.ReadFile("../runner/testdata/sleep.tf.json")
	if err != nil {
		t.Error(err)
	}

	plan["sleep.tf.json"] = uglyJson(lBytes)

	vBytes, err := ioutil.ReadFile("../tmpl/testdata/vars.json")
	if err != nil {
		t.Error(err)
	}

	pBytes, _ := json.Marshal(plan)
	vBytes = uglyJson(vBytes)

	t.Run("Should create a layout in the workspace", func(t *testing.T) {
		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: workspaceId, Plan: pBytes}
		resp, err := server.SaveLayout(context.Background(), req)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, resp, &Ok{})
	})

	t.Run("Layout with provider conflict without worksapce should not error", func(t *testing.T) {
		id := fmt.Sprintf("workspace-conflict")
		wreq := &SaveWorkspaceRequest{Id: id}
		if _, err := server.SaveWorkspace(context.Background(), wreq); err != nil {
			errors.Wrap(err, Errors_INVALID_VALUE.String())
		}

		plan2 := map[string]json.RawMessage{}
		fBytes, err := ioutil.ReadFile("./testdata/file.tf.json")
		if err != nil {
			t.Error(err)
		}

		plan2["file.tf.json"] = uglyJson(fBytes)
		pBytes, _ := json.Marshal(plan2)
		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: id, Plan: pBytes}

		if _, err = server.SaveLayout(context.Background(), req); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Layout with provider conflict without workspace should error", func(t *testing.T) {
		id := fmt.Sprintf("workspace-conflict")
		wv := &types.Vars{"aws": nil}
		wvars, _ := json.Marshal(wv)
		wreq := &SaveWorkspaceRequest{Id: id, Providers: wvars}
		if _, err := server.SaveWorkspace(context.Background(), wreq); err != nil {
			errors.Wrap(err, Errors_INVALID_VALUE.String())
		}

		plan2 := map[string]json.RawMessage{}
		fBytes, err := ioutil.ReadFile("./testdata/file.tf.json")
		if err != nil {
			t.Error(err)
		}

		plan2["file.tf.json"] = uglyJson(fBytes)
		pBytes, _ := json.Marshal(plan2)
		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: id, Plan: pBytes}

		_, err = server.SaveLayout(context.Background(), req)
		if err == nil {
			t.Fatal("Should have complained about a conflicting Provider")
		}
	})

	t.Run("Should get the layout that was created", func(t *testing.T) {
		req := &LayoutRequest{WorkspaceId: workspaceId, Id: layoutId}
		resp, err := server.GetLayout(context.Background(), req)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, resp.Id, layoutId)
		assert.Equal(t, resp.Status, Status_INACTIVE)
		assert.Equal(t, resp.Workspaceid, workspaceId)
		assert.Equal(t, resp.Plan, pBytes)
	})

	t.Run("Should save a watch", func(t *testing.T) {
		req := &StartWatchRequest{
			WorkspaceId:     workspaceId,
			Id:              layoutId,
			SuccessCallback: "http://google.com",
			FailureCallback: "http://yahoo.com",
		}

		resp, err := server.StartWatch(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, resp, &Ok{})
	})

	t.Run("Should unset a watch", func(t *testing.T) {
		req := &StopWatchRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId,
		}

		resp, err := server.StopWatch(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, resp, &Ok{})
	})

	t.Run("Should apply a layout", func(t *testing.T) {
		req := &ApplyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId,
			Dry:         true,
			Vars:        vBytes,
		}

		resp, err := server.ApplyLayout(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, JobState_PENDING, resp.Status)
		assert.NotEmpty(t, resp.Id)

		assert.Equal(t, jobQueue.Store, []string{resp.Id})

		job := types.Job{Id: resp.Id, LayoutId: layoutId}
		tree := types.MakeTree(workspaceId)
		if err := store.Get(&job, tree); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, layoutId, job.LayoutId)
		assert.Equal(t, int32(JobState_PENDING), job.Status)
		assert.Equal(t, int32(Operation_APPLY), job.Op)
		assert.Equal(t, true, job.Dry)
		assert.NotEmpty(t, job.LayoutVersion)
	})

	lockKey := fmt.Sprintf("%v-%v", workspaceId, layoutId)

	t.Run("Should Lock a run till completed by worker", func(t *testing.T) {
		req := &ApplyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId,
			Dry:         true,
			Vars:        vBytes,
		}

		_, err := server.ApplyLayout(context.Background(), req)
		if err == nil {
			t.Fatalf("Should have failed with a Lock, key: %s-%s", workspaceId, layoutId)
		}
	})

	t.Run("Should allow unlocking a Layout", func(t *testing.T) {
		if err := store.Unlock(lockKey); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Unlocking is Idempotent", func(t *testing.T) {
		if err := store.Unlock(lockKey); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("Should Destroy a layout", func(t *testing.T) {
		req := &DestroyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId,
		}

		resp, err := server.DestroyLayout(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, JobState_PENDING, resp.Status)
		assert.NotEmpty(t, resp.Id)

		assert.Equal(t, jobQueue.Store[len(jobQueue.Store)-1], resp.Id, fmt.Sprintf("%v", jobQueue.Store))

		job := types.Job{Id: resp.Id, LayoutId: layoutId}
		tree := types.MakeTree(workspaceId)
		if err := store.Get(&job, tree); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, layoutId, job.LayoutId)
		assert.Equal(t, int32(JobState_PENDING), job.Status)
		assert.Equal(t, int32(Operation_DESTROY), job.Op)
		assert.Equal(t, false, job.Dry)
		assert.NotEmpty(t, job.LayoutVersion)
	})
}
