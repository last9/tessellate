package storage

import (
	"log"
	"testing"

	"math/rand"
	"time"

	"github.com/tsocial/tessellate/storage/consul"
	"github.com/tsocial/tessellate/storage/types"
	"github.com/stretchr/testify/assert"
	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"os"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"github.com/tsocial/tessellate/server"
)

var store Storer

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Deletes all the keys in the prefix / on Consul.
func deleteTree(client *api.Client) error {
	client.KV().Put(&api.KVPair{}, &api.WriteOptions{})

	if _, err := client.KV().DeleteTree("testing/", &api.WriteOptions{}); err != nil {
		return errors.Wrap(err, "Cannot delete all keys under prefix /")
	}

	return nil
}

func TestMain(m *testing.M) {
	//Seed Random number generator.
	rand.Seed(time.Now().UnixNano())

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	store = consul.MakeConsulStore("127.0.0.1:8500")
	store.Setup()
/*
	os.Exit(func () int {
		defer deleteTree(store.GetClient())

		y := m.Run()
		return y
	}())
	*/
	os.Exit(m.Run())
}


func TestStorer(t *testing.T) {
	workspace_id := "testing/workspace-1"
	layout_id := "layout-1"

	t.Run("Workspace tests", func(t *testing.T) {
		t.Run("Workspace does not exist", func(t *testing.T) {
			v, err := store.GetWorkspace(workspace_id)
			if err == nil {
				t.Fatal("Should have failed with an Error")
			}
			assert.Nil(t, v)
		})

		t.Run("Get a Workspace after creation", func(t *testing.T) {
			err := store.SaveWorkspace(workspace_id, nil)
			if err != nil {
				t.Fatal(err)
			}

			val, err := store.GetWorkspace(workspace_id)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "latest", val.Version)
		})

		t.Run("Resaving a Workspace doesn't raise an Error", func(t *testing.T) {
			vars := types.Vars(map[string]interface{}{"key": 1})
			if err := store.SaveWorkspace(workspace_id, &vars); err != nil {
				t.Fatal(err)
			}

			val, err := store.GetWorkspace(workspace_id)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "latest", val.Version)
		})
	})

	t.Run("Layout tests", func(t *testing.T) {
		layouts, err := getLayouts()
		if err != nil {
			t.Fatal(err)
			return
		}

		t.Run("Save Layouts within workspace test", func(t *testing.T) {
			store.SaveLayout(workspace_id, layout_id, layouts, nil)
		})

		t.Run(" Get the layout for the mentioned workspace and name", func(t *testing.T) {
			layout, _ := store.GetLayout(workspace_id, layout_id)
			fmt.Println(layout.Plan)
		})

		t.Run("Get all the layouts and all the versioned plans in the given layout", func(t *testing.T) {
			layouts, _ := store.GetWorkspaceLayouts(workspace_id)
			for k, v := range layouts {
				fmt.Println(k, v)
			}
		})

		t.Run("Set layout status to inactive", func(t *testing.T) {
			store.SetLayoutStatus(workspace_id, layout_id, server.Status_INACTIVE.String())
		})

		t.Run("Get the layout status which should be inactive", func(t *testing.T) {
			status, _ := store.GetLayoutStatus(workspace_id, layout_id)

			assert.Equal(t, server.Status_INACTIVE.String(), status)
		})
	})

}

// read tf.json files and form a map of key-value
func getLayouts() (map[string]interface{}, error) {
	layouts := make(map[string]interface{}, 2)

	l := json.RawMessage{}
	d, err := ioutil.ReadFile("../runner/testdata/sleep.tf.json")
	if err != nil {
		return nil, errors.Wrap(err, server.Errors_INTERNAL_ERROR.String())
	}

	v, err := ioutil.ReadFile("../runner/testdata/vars.json")
	if err != nil {
		return nil, errors.Wrap(err, server.Errors_INTERNAL_ERROR.String())
	}

	if err := json.Unmarshal(d, &l); err != nil {
		return nil, errors.Wrap(err, server.Errors_INVALID_VALUE.String())
	}

	l = json.RawMessage{}
	if err := json.Unmarshal(v, &l); err != nil {
		return nil, errors.Wrap(err, server.Errors_INVALID_VALUE.String())
	}

	layouts["sleep"] = l
	layouts["env"] = v

	return layouts, nil
}