package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/server"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestGetAllWorkspace(t *testing.T) {

	t.Run("Should return map with name, version and all provider names", func(t *testing.T) {
		vars := `{
			"provider": [
				{
					"alicloud": {
					}
				},
				{
					"aws": {
						"access_key": "***********",
						"region": "ap-south-1",
						"secret_key": "***********"
					}
				}
			],
			"variable": {
				"region": {}
			}
		}`
		w := server.Workspace{Name: "W1", Vars: []byte(vars), Version: "latest"}
		out := workspaceMap(&w)
		assert.Equal(t, out["Name"], "W1")
		assert.Equal(t, out["Version"], "latest")
		assert.Equal(t, out["Providers"], "alicloud, aws")
	})

	t.Run("Should return map with name and version when provider key is not available", func(t *testing.T) {
		vars := `{
			"variable": {
				"region": {}
			}
		}`
		w := server.Workspace{Name: "W2", Vars: []byte(vars), Version: "latest"}
		out := workspaceMap(&w)
		assert.Equal(t, out["Name"], "W2")
		assert.Equal(t, out["Version"], "latest")
		assert.Equal(t, out["Providers"], nil)
	})

	t.Run("Should return map with name and version when provider list is empty", func(t *testing.T) {
		vars := `{
			"provider": [],
			"variable": {
				"region": {}
			}
		}`
		w := server.Workspace{Name: "W3", Vars: []byte(vars), Version: "latest"}
		out := workspaceMap(&w)
		assert.Equal(t, out["Name"], "W3")
		assert.Equal(t, out["Version"], "latest")
		assert.Equal(t, out["Providers"], nil)
	})

	t.Run("Should return map with name and version when vars object is empty", func(t *testing.T) {
		vars := `{
			"provider": [],
			"variable": {
				"region": {}
			}
		}`
		w := server.Workspace{Name: "W4", Vars: []byte(vars), Version: "latest"}
		out := workspaceMap(&w)
		assert.Equal(t, out["Name"], "W4")
		assert.Equal(t, out["Version"], "latest")
		assert.Equal(t, out["Providers"], nil)
	})
}
