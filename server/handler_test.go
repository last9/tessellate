package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

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
		_, err := server.SaveWorkspace(context.Background(), req)
		assert.Nil(t, err, Errors_INVALID_VALUE.String())
	})

	t.Run("Should get the same workspace that was created.", func(t *testing.T) {
		req := &GetWorkspaceRequest{Id: id}
		resp, err := server.GetWorkspace(context.Background(), req)
		assert.Nil(t, err, Errors_INVALID_VALUE.String())
		assert.Equal(t, id, resp.Name)
	})
}

func TestServer_SaveAndGetLayout(t *testing.T) {
	workspaceId := fmt.Sprintf("workspace-%s", utils.RandString(8))
	layoutId := fmt.Sprintf("layout-%s", utils.RandString(8))

	jobQueue := dispatcher.NewInMemory()
	dispatcher.Set(jobQueue)

	plan := map[string]json.RawMessage{}

	lBytes, err := ioutil.ReadFile("../runner/testdata/sleep.tf.json")
	assert.Nil(t, err)

	plan["sleep.tf.json"] = uglyJson(lBytes)

	vBytes, err := ioutil.ReadFile("../tmpl/testdata/vars.json")
	assert.Nil(t, err)

	pBytes, _ := json.Marshal(plan)
	vBytes = uglyJson(vBytes)

	t.Run("Should create a layout in the workspace", func(t *testing.T) {
		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: workspaceId, Plan: pBytes}
		resp, err := server.SaveLayout(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, resp.LayoutId, layoutId)

		//get saved layout and match content
		getReq := &LayoutRequest{WorkspaceId: workspaceId, Id: layoutId}
		gResp, err := server.GetLayout(context.Background(), getReq)
		assert.Nil(t, err)
		assert.Equal(t, gResp.Plan, pBytes)
		assert.Equal(t, gResp.Id, layoutId)
		assert.Equal(t, gResp.Workspaceid, workspaceId)
	})

	t.Run("Layout with provider conflict without worksapce should not error", func(t *testing.T) {
		id := fmt.Sprintf("workspace-conflict")
		wReq := &SaveWorkspaceRequest{Id: id}
		_, err := server.SaveWorkspace(context.Background(), wReq)
		assert.Nil(t, err, Errors_INVALID_VALUE.String())

		plan2 := map[string]json.RawMessage{}
		fBytes, err := ioutil.ReadFile("./testdata/file.tf.json")
		assert.Nil(t, err)

		plan2["file.tf.json"] = uglyJson(fBytes)
		pBytes, _ := json.Marshal(plan2)
		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: id, Plan: pBytes}

		_, err = server.SaveLayout(context.Background(), req)
		assert.Nil(t, err)
	})

	t.Run("Layout with provider conflict without workspace should error", func(t *testing.T) {
		id := fmt.Sprintf("workspace-conflict")
		wv := &types.Vars{"aws": nil}
		wVars, _ := json.Marshal(wv)
		wReq := &SaveWorkspaceRequest{Id: id, Providers: wVars}
		_, err := server.SaveWorkspace(context.Background(), wReq)
		assert.Nil(t, err, Errors_INVALID_VALUE.String())

		plan2 := map[string]json.RawMessage{}
		fBytes, err := ioutil.ReadFile("./testdata/file.tf.json")
		assert.Nil(t, err)

		plan2["file.tf.json"] = uglyJson(fBytes)
		pBytes, _ := json.Marshal(plan2)
		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: id, Plan: pBytes}

		_, err = server.SaveLayout(context.Background(), req)
		assert.NotNil(t, err, "Should have complained about a conflicting Provider")
	})

	t.Run("Should get the layout that was created", func(t *testing.T) {
		req := &LayoutRequest{WorkspaceId: workspaceId, Id: layoutId}
		resp, err := server.GetLayout(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, resp.Id, layoutId)
		assert.Equal(t, resp.Status, Status_INACTIVE)
		assert.Equal(t, resp.Workspaceid, workspaceId)
		assert.Equal(t, resp.Plan, pBytes)
	})

	t.Run("Should get all the layout that was created", func(t *testing.T) {
		wid := workspaceId
		lid := "l1"

		req := &SaveLayoutRequest{Id: lid, WorkspaceId: wid, Plan: pBytes}
		resp, err := server.SaveLayout(context.Background(), req)
		assert.Nil(t, err)

		assert.Equal(t, resp.LayoutId, lid)

		//get saved layout and match content
		getReq := &LayoutRequest{WorkspaceId: wid, Id: lid}
		gResp, err := server.GetLayout(context.Background(), getReq)
		assert.Nil(t, err)
		assert.Equal(t, pBytes, gResp.Plan)
		assert.Equal(t, lid, gResp.Id)
		assert.Equal(t, wid, gResp.Workspaceid)

		nReq := &GetWorkspaceLayoutsRequest{Id: workspaceId}
		nResp, err := server.GetWorkspaceLayouts(context.Background(), nReq)
		assert.Nil(t, err)
		layouts := []*Layout{&Layout{Workspaceid: workspaceId, Id: lid},
			&Layout{Workspaceid: workspaceId, Id: layoutId}}
		assert.Equal(t, 2, len(nResp.Layouts))
		assert.ElementsMatch(t, layouts, nResp.Layouts)
	})

	t.Run("Should get empty layout list when workspace not exist", func(t *testing.T) {
		nReq := &GetWorkspaceLayoutsRequest{Id: "fakeworkspace"}
		nResp, err := server.GetWorkspaceLayouts(context.Background(), nReq)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(nResp.Layouts))
	})

	t.Run("Should save a watch", func(t *testing.T) {
		req := &StartWatchRequest{
			WorkspaceId:     workspaceId,
			Id:              layoutId,
			SuccessCallback: "http://google.com",
			FailureCallback: "http://yahoo.com",
		}

		resp, err := server.StartWatch(context.Background(), req)
		assert.Nil(t, err)
		assert.Equal(t, resp, &Ok{})
	})

	t.Run("Should unset a watch", func(t *testing.T) {
		req := &StopWatchRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId,
		}

		resp, err := server.StopWatch(context.Background(), req)
		assert.Nil(t, err)
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
		assert.Nil(t, err)

		assert.Equal(t, JobState_PENDING, resp.Status)
		assert.NotEmpty(t, resp.Id)

		assert.Equal(t, jobQueue.Store, []string{resp.Id})

		job := types.Job{Id: resp.Id, LayoutId: layoutId}
		tree := types.MakeTree(workspaceId)
		err = store.Get(&job, tree)
		assert.Nil(t, err)

		assert.Equal(t, layoutId, job.LayoutId)
		assert.Equal(t, int32(JobState_PENDING), job.Status)
		assert.Equal(t, int32(Operation_APPLY), job.Op)
		assert.Equal(t, true, job.Dry)
		assert.NotEmpty(t, job.LayoutVersion)
		assert.Equal(t, int64(0), job.Retry)
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
		assert.NotNil(t, err, fmt.Sprintf("Should have failed with a Lock, key: %s-%s", workspaceId, layoutId))
	})

	t.Run("Should allow unlocking a Layout", func(t *testing.T) {
		err := store.Unlock(lockKey)
		assert.Nil(t, err)
	})

	t.Run("Unlocking is Idempotent", func(t *testing.T) {
		err := store.Unlock(lockKey)
		assert.Nil(t, err)
	})

	t.Run("Should Destroy a layout", func(t *testing.T) {
		req := &DestroyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId,
		}

		resp, err := server.DestroyLayout(context.Background(), req)
		assert.Nil(t, err)

		assert.Equal(t, JobState_PENDING, resp.Status)
		assert.NotEmpty(t, resp.Id)

		assert.Equal(t, jobQueue.Store[len(jobQueue.Store)-1], resp.Id, fmt.Sprintf("%v", jobQueue.Store))

		job := types.Job{Id: resp.Id, LayoutId: layoutId}
		tree := types.MakeTree(workspaceId)
		err = store.Get(&job, tree)
		assert.Nil(t, err)

		assert.Equal(t, layoutId, job.LayoutId)
		assert.Equal(t, int32(JobState_PENDING), job.Status)
		assert.Equal(t, int32(Operation_DESTROY), job.Op)
		assert.Equal(t, false, job.Dry)
		assert.Equal(t, int64(0), job.Retry)
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
	assert.Nil(t, err)

	plan["sleep.tf.json"] = uglyJson(lBytes)

	vBytes, err := ioutil.ReadFile("../tmpl/testdata/vars.json")
	assert.Nil(t, err)

	pBytes, _ := json.Marshal(plan)
	vBytes = uglyJson(vBytes)

	t.Run("Should create a layout in the workspace with dry flag with empty state", func(t *testing.T) {
		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: workspaceId, Plan: pBytes, Dry: true}
		resp, err := server.SaveLayout(context.Background(), req)
		assert.Nil(t, err)
		newLayoutId := layoutId + drySuffix
		assert.Equal(t, newLayoutId, resp.LayoutId)

		tree := types.MakeTree(workspaceId)
		l := types.Layout{
			Id:   newLayoutId,
			Plan: map[string]json.RawMessage{},
		}

		vAfterSave, err := store.GetVersions(&l, tree)
		assert.Nil(t, err)

		assert.Equal(t, 2, len(vAfterSave))

		//get saved layout and match content
		getReq := &LayoutRequest{WorkspaceId: workspaceId, Id: newLayoutId}
		gResp, err := server.GetLayout(context.Background(), getReq)
		assert.Nil(t, err)
		assert.Equal(t, gResp.Plan, pBytes)
		assert.Equal(t, gResp.Id, newLayoutId)
		assert.Equal(t, gResp.Workspaceid, workspaceId)
	})

	t.Run("Should create a layout in the workspace with dry flag with existing state", func(t *testing.T) {
		//temporary saving state
		key := path.Join(state, workspaceId, layoutId)
		value := "some test value"
		err := store.SaveKey(key, []byte(value))
		assert.Nil(t, err)

		tree := types.MakeTree(workspaceId)
		l := types.Layout{
			Id:   layoutId,
			Plan: map[string]json.RawMessage{},
		}

		vBeforeSave, err := store.GetVersions(&l, tree)
		assert.Nil(t, err)

		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: workspaceId, Plan: pBytes, Dry: true}
		resp, err := server.SaveLayout(context.Background(), req)

		assert.Nil(t, err)
		newLayoutId := layoutId + drySuffix
		assert.Equal(t, resp.LayoutId, newLayoutId)

		key = path.Join(state, workspaceId, newLayoutId)
		newValue, err := store.GetKey(key)
		assert.Nil(t, err)
		assert.Equal(t, value, string(newValue))

		vAfterSave, err := store.GetVersions(&l, tree)
		assert.Nil(t, err)

		assert.Equal(t, len(vBeforeSave), len(vAfterSave))
	})

	t.Run("Should not apply a dry layout without dry flag", func(t *testing.T) {
		req := &ApplyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId + drySuffix,
			Dry:         false,
			Vars:        vBytes,
		}

		_, err := server.ApplyLayout(context.Background(), req)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Sprintf("Operation not allowed, on %s, use --dry to run a terraform plan",
			layoutId+drySuffix), err.Error())
	})

	t.Run("Should apply a layout", func(t *testing.T) {
		newLayoutId := layoutId + drySuffix
		req := &ApplyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          newLayoutId,
			Dry:         true,
			Vars:        vBytes,
		}
		resp, err := server.ApplyLayout(context.Background(), req)
		assert.Nil(t, err)

		assert.Equal(t, JobState_PENDING, resp.Status)
		assert.NotEmpty(t, resp.Id)

		assert.Equal(t, jobQueue.Store, []string{resp.Id})

		job := types.Job{Id: resp.Id, LayoutId: newLayoutId}
		tree := types.MakeTree(workspaceId)
		err = store.Get(&job, tree)
		assert.Nil(t, err)

		assert.Equal(t, newLayoutId, job.LayoutId)
		assert.Equal(t, int32(JobState_PENDING), job.Status)
		assert.Equal(t, int32(Operation_APPLY), job.Op)
		assert.Equal(t, true, job.Dry)
		assert.NotEmpty(t, job.LayoutVersion)
		assert.Equal(t, int64(0), job.Retry)

		lockKey := fmt.Sprintf("%v-%v", workspaceId, newLayoutId)
		err = store.Unlock(lockKey)
		assert.Nil(t, err)
	})

	t.Run("Should destroy a layout", func(t *testing.T) {
		newLayoutId := layoutId + drySuffix
		req := &DestroyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          newLayoutId,
			Vars:        vBytes,
		}

		resp, err := server.DestroyLayout(context.Background(), req)
		assert.Nil(t, err)

		assert.Equal(t, JobState_PENDING, resp.Status)
		assert.NotEmpty(t, resp.Id)

		assert.Equal(t, jobQueue.Store[len(jobQueue.Store)-1], resp.Id, fmt.Sprintf("%v", jobQueue.Store))

		job := types.Job{Id: resp.Id, LayoutId: newLayoutId}
		tree := types.MakeTree(workspaceId)
		err = store.Get(&job, tree)
		assert.Nil(t, err)

		assert.Equal(t, newLayoutId, job.LayoutId)
		assert.Equal(t, int32(JobState_PENDING), job.Status)
		assert.Equal(t, int32(Operation_DESTROY), job.Op)
		assert.Equal(t, false, job.Dry)
		assert.Equal(t, int64(0), job.Retry)
		assert.NotEmpty(t, job.LayoutVersion)
	})
}

func TestServer_GetOutput(t *testing.T) {
	workspace := uuid.NewV4().String()
	layout := uuid.NewV4().String()

	key := fmt.Sprintf("state/%s/%s", workspace, layout)
	valBytes, err := ioutil.ReadFile("./testdata/output.tfstate")
	assert.Nil(t, err)

	err = store.SaveKey(key, valBytes)
	assert.Nil(t, err)

	req := &GetOutputRequest{
		LayoutId:    layout,
		WorkspaceId: workspace,
	}

	resp, err := server.GetOutput(context.Background(), req)
	assert.Nil(t, err)

	s := StateStruct{}
	err = json.Unmarshal(valBytes, &s)
	assert.Nil(t, err)

	expected := map[string]interface{}{}
	for k := range s.Modules[0].Outputs {
		expected[k] = s.Modules[0].Outputs[k].Value
	}

	outMap := map[string]interface{}{}
	err = json.Unmarshal(resp.Output, &outMap)
	assert.Nil(t, err)

	assert.Equal(t, expected, outMap)
}

func TestServer_SaveApplyAndDestroyLayoutWithRetry(t *testing.T) {
	workspaceId := fmt.Sprintf("workspace-%s", utils.RandString(8))
	layoutId := fmt.Sprintf("layout-%s", utils.RandString(8))

	jobQueue := dispatcher.NewInMemory()
	dispatcher.Set(jobQueue)

	plan := map[string]json.RawMessage{}

	lBytes, err := ioutil.ReadFile("../runner/testdata/sleep.tf.json")
	assert.Nil(t, err)

	plan["sleep.tf.json"] = uglyJson(lBytes)

	vBytes, err := ioutil.ReadFile("../tmpl/testdata/vars.json")
	assert.Nil(t, err)

	pBytes, _ := json.Marshal(plan)
	vBytes = uglyJson(vBytes)

	t.Run("Should create a layout in the workspace", func(t *testing.T) {
		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: workspaceId, Plan: pBytes}
		resp, err := server.SaveLayout(context.Background(), req)

		assert.Nil(t, err)

		assert.Equal(t, resp.LayoutId, layoutId)

		//get saved layout and match content
		getReq := &LayoutRequest{WorkspaceId: workspaceId, Id: layoutId}
		gResp, err := server.GetLayout(context.Background(), getReq)
		assert.Nil(t, err)
		assert.Equal(t, gResp.Plan, pBytes)
		assert.Equal(t, gResp.Id, layoutId)
		assert.Equal(t, gResp.Workspaceid, workspaceId)
	})

	t.Run("Should apply a layout with retry 4", func(t *testing.T) {
		req := &ApplyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId,
			Dry:         false,
			Vars:        vBytes,
			Retry:       4,
		}

		resp, err := server.ApplyLayout(context.Background(), req)
		assert.Nil(t, err)

		assert.Equal(t, JobState_PENDING, resp.Status)
		assert.NotEmpty(t, resp.Id)

		assert.Equal(t, jobQueue.Store, []string{resp.Id})

		job := types.Job{Id: resp.Id, LayoutId: layoutId}
		tree := types.MakeTree(workspaceId)
		err = store.Get(&job, tree)
		assert.Nil(t, err)

		assert.Equal(t, layoutId, job.LayoutId)
		assert.Equal(t, int32(JobState_PENDING), job.Status)
		assert.Equal(t, int32(Operation_APPLY), job.Op)
		assert.Equal(t, false, job.Dry)
		assert.NotEmpty(t, job.LayoutVersion)
		assert.Equal(t, int64(4), job.Retry)

		lockKey := fmt.Sprintf("%v-%v", workspaceId, layoutId)
		err = store.Unlock(lockKey)
		assert.Nil(t, err)
	})

	t.Run("Should raise validation error when `retry=<nagetive_value>` while applying a layout", func(t *testing.T) {
		req := &ApplyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId,
			Dry:         false,
			Vars:        vBytes,
			Retry:       -3,
		}

		_, err := server.ApplyLayout(context.Background(), req)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "ApplyLayoutRequest.Retry: value must be greater than or equal to 0")
	})

	t.Run("Should raise validation error when `retry=<nagetive_value>` while destroying a layout", func(t *testing.T) {
		req := &DestroyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId,
			Vars:        vBytes,
			Retry:       -5,
		}

		_, err := server.DestroyLayout(context.Background(), req)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "DestroyLayoutRequest.Retry: value must be greater than or equal to 0")
	})

	t.Run("Should destroy a layout with retry 6", func(t *testing.T) {
		req := &DestroyLayoutRequest{
			WorkspaceId: workspaceId,
			Id:          layoutId,
			Vars:        vBytes,
			Retry:       6,
		}

		resp, err := server.DestroyLayout(context.Background(), req)
		assert.Nil(t, err)

		assert.Equal(t, JobState_PENDING, resp.Status)
		assert.NotEmpty(t, resp.Id)

		assert.Equal(t, jobQueue.Store[len(jobQueue.Store)-1], resp.Id, fmt.Sprintf("%v", jobQueue.Store))

		job := types.Job{Id: resp.Id, LayoutId: layoutId}
		tree := types.MakeTree(workspaceId)
		err = store.Get(&job, tree)
		assert.Nil(t, err)

		assert.Equal(t, layoutId, job.LayoutId)
		assert.Equal(t, int32(JobState_PENDING), job.Status)
		assert.Equal(t, int32(Operation_DESTROY), job.Op)
		assert.Equal(t, false, job.Dry)
		assert.Equal(t, int64(6), job.Retry)
		assert.NotEmpty(t, job.LayoutVersion)
	})
}
