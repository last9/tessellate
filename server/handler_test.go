package server

import (
	"os"
	"testing"

	"github.com/tsocial/tessellate/storage/consul"

	"github.com/pkg/errors"
	"fmt"
	"context"
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
