package server

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/tsocial/tessellate/storage/consul"

	"context"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var server TessellateServer

func TestMain(m *testing.M) {

	store := consul.MakeConsulStore()
	store.Setup()

	server = New(store)

	os.Exit(m.Run())
}

func uglyJson(b []byte) []byte {
	var t map[string]interface{}
	json.Unmarshal(b, &t)
	b2, _ := json.Marshal(t)
	return b2
}

func TestServer_SaveAndGetWorkspace(t *testing.T) {
	id := "workspace-1"

	t.Run("Should save a workspace.", func(t *testing.T) {
		req := &SaveWorkspaceRequest{Id: id}
		resp, err := server.SaveWorkspace(context.Background(), req)

		if err != nil {
			errors.Wrap(err, Errors_INVALID_VALUE.String())
		}
		fmt.Print(resp.String())
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
	workspaceId := "workspace-1"
	layoutId := "layout-1"
	plan := map[string][]byte{}

	lBytes, err := ioutil.ReadFile("../runner/testdata/sleep.tf.json")
	if err != nil {
		t.Error(err)
	}

	plan["sleep"] = uglyJson(lBytes)

	vBytes, err := ioutil.ReadFile("../runner/testdata/vars.json")
	if err != nil {
		t.Error(err)
	}

	vBytes = uglyJson(vBytes)

	t.Run("Should create a layout in the workspace", func(t *testing.T) {
		req := &SaveLayoutRequest{Id: layoutId, WorkspaceId: workspaceId, Plan: plan, Vars: vBytes}
		resp, err := server.SaveLayout(context.Background(), req)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, resp, &Ok{})
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
		assert.Equal(t, resp.Plan, plan)

		t.Log(string(resp.Vars))
		t.Log(string(vBytes))

		assert.Equal(t, resp.Vars, vBytes)
	})
}
