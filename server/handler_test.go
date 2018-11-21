package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"path"
	"testing"

	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/dispatcher"
	"github.com/tsocial/tessellate/storage"
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

		assert.Equal(t, resp.LayoutId, layoutId)

		//get saved layout and match content
		getReq := &LayoutRequest{WorkspaceId: workspaceId, Id: layoutId}
		gResp, err := server.GetLayout(context.Background(), getReq)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, gResp.Plan, pBytes)
		assert.Equal(t, gResp.Id, layoutId)
		assert.Equal(t, gResp.Workspaceid, workspaceId)
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

	t.Run("Should get all the layout that was created", func(t *testing.T) {
		req := &GetWorkspaceLayoutsRequest{Id: workspaceId}
		resp, err := server.GetWorkspaceLayouts(context.Background(), req)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(resp.Layouts), 1)
		assert.Equal(t, resp.Layouts[0].Id, layoutId)
		assert.Equal(t, resp.Layouts[0].Workspaceid, workspaceId)
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
			Retry:       3,
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

func TestServer_SaveAndGetLayout_Dry(t *testing.T) {
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

	t.Run("Should create a layout in the workspace with dry flag with empty state", func(t *testing.T) {
		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: workspaceId, Plan: pBytes, Dry: true}
		resp, err := server.SaveLayout(context.Background(), req)

		if err != nil {
			t.Fatal(err)
		}
		newLayoutId := layoutId + drysuffix
		assert.Equal(t, newLayoutId, resp.LayoutId)

		tree := types.MakeTree(workspaceId)
		l := types.Layout{
			Id:   newLayoutId,
			Plan: map[string]json.RawMessage{},
		}

		vAfterSave, err := store.GetVersions(&l, tree)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, len(vAfterSave))

		//get saved layout and match content
		getReq := &LayoutRequest{WorkspaceId: workspaceId, Id: newLayoutId}
		gResp, err := server.GetLayout(context.Background(), getReq)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, gResp.Plan, pBytes)
		assert.Equal(t, gResp.Id, newLayoutId)
		assert.Equal(t, gResp.Workspaceid, workspaceId)
	})

	t.Run("Should create a layout in the workspace with dry flag with existing state", func(t *testing.T) {
		//temporary saving state
		key := path.Join(state, workspaceId, layoutId)
		value := "some test value"
		store.SaveKey(key, []byte(value))

		tree := types.MakeTree(workspaceId)
		l := types.Layout{
			Id:   layoutId,
			Plan: map[string]json.RawMessage{},
		}

		vBeforeSave, err := store.GetVersions(&l, tree)
		if err != nil {
			t.Fatal(err)
		}

		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: workspaceId, Plan: pBytes, Dry: true}
		resp, err := server.SaveLayout(context.Background(), req)

		if err != nil {
			t.Fatal(err)
		}
		newLayoutId := layoutId + drysuffix
		assert.Equal(t, resp.LayoutId, newLayoutId)

		key = path.Join(state, workspaceId, newLayoutId)
		newvalue, err := store.GetKey(key)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, value, string(newvalue))

		vAfterSave, err := store.GetVersions(&l, tree)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, len(vBeforeSave), len(vAfterSave))
	})

	t.Run("Should not apply a dry layout without dry flag", func(t *testing.T) {
		req := &ApplyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId + drysuffix,
			Dry:         false,
			Vars:        vBytes,
		}

		_, err := server.ApplyLayout(context.Background(), req)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Sprintf("Operation not allowed, on %s, use --dry to run a terraform plan",
			layoutId+drysuffix), err.Error())
	})
}

func TestServer_GetOutput(t *testing.T) {
	workspace := uuid.NewV4().String()
	layout := uuid.NewV4().String()

	key := fmt.Sprintf("state/%s/%s", workspace, layout)
	valBytes, err := ioutil.ReadFile("./testdata/output.tfstate")
	if err != nil {
		t.Fatal(err)
	}

	if err := store.SaveKey(key, valBytes); err != nil {
		t.Fatal(err)
	}

	req := &GetOutputRequest{
		LayoutId:    layout,
		WorkspaceId: workspace,
	}

	resp, err := server.GetOutput(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	s := StateStruct{}
	if err := json.Unmarshal(valBytes, &s); err != nil {
		t.Fatal(err)
	}

	expected := map[string]interface{}{}
	for k := range s.Modules[0].Outputs {
		expected[k] = s.Modules[0].Outputs[k].Value
	}

	outMap := map[string]interface{}{}
	if err := json.Unmarshal(resp.Output, &outMap); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected, outMap)
}
