package server

import (
	"os"
	"testing"

	"github.com/tsocial/tessellate/storage/consul"

	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
)

var server TessellateServer

func TestMain(m *testing.M) {

	store := consul.MakeConsulStore()
	store.Setup()

	server = New(store)

	os.Exit(m.Run())
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

func TestServer_SaveLayout(t *testing.T) {
	workspace_id := "workspace-1"
	layout_id := "layout-1"
	plan := make(map[string][]byte, 2)

	l := json.RawMessage{}
	d, err := ioutil.ReadFile("../runner/testdata/sleep.tf.json")
	if err != nil {
		t.Error(err)
	}

	v, err := ioutil.ReadFile("../tmpl/testdata/vars.json")
	if err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(d, &l); err != nil {
		t.Fatal(err)
		return
	}

	l = json.RawMessage{}
	if err := json.Unmarshal(v, &l); err != nil {
		t.Fatal(err)
		return
	}

	plan["sleep"] = l
	plan["env"] = l

	t.Run("Should create a layout in the workspace", func(t *testing.T) {
		req := &SaveLayoutRequest{Id: layout_id, WorkspaceId: workspace_id, Plan: plan}
		resp, err := server.SaveLayout(context.Background(), req)

		if err != nil {
			errors.Wrap(err, Errors_INVALID_VALUE.String())
		}
		fmt.Print(resp.String())
	})
}
