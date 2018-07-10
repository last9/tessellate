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

	os.Exit(func () int {
		defer deleteTree(store.GetClient())

		y := m.Run()
		return y
	}())
}

func TestStorer(t *testing.T) {
	workspace := "testing/hello"

	t.Run("Workspace tests", func(t *testing.T) {
		t.Run("Workspace does not exist", func(t *testing.T) {
			v, err := store.GetWorkspace(workspace)
			if err == nil {
				t.Fatal("Should have failed with an Error")
			}
			assert.Nil(t, v)
		})

		t.Run("Get an Workspace after creation", func(t *testing.T) {
			err := store.SaveWorkspace(workspace, nil)
			if err != nil {
				t.Fatal(err)
			}

			val, err := store.GetWorkspace(workspace)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "latest", val.Version)
		})

		t.Run("Resaving a Workspace doesn't raise Error", func(t *testing.T) {
			vars := types.Vars(map[string]interface{}{"key": 1})
			if err := store.SaveWorkspace(workspace, &vars); err != nil {
				t.Fatal(err)
			}

			val, err := store.GetWorkspace(workspace)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, "latest", val.Version)
		})
	})



	//
	//t.Run("Vars tests", func(t *testing.T) {
	//	d := map[string]interface{}{"ok": "1", "world": "2"}
	//
	//	t.Run("Get Vars that don't exist", func(t *testing.T) {
	//		_, err := store.GetVars(workspace, "var1")
	//		assert.Error(t, err)
	//		assert.Contains(t, err.Error(), "key has not been set")
	//	})
	//
	//	t.Run("Save Vars", func(t *testing.T) {
	//		err := store.SaveVars(workspace, "var1", d)
	//		assert.NoError(t, err)
	//	})
	//
	//	t.Run("Get saved vars", func(t *testing.T) {
	//		out, err := store.GetVars(workspace, "var1")
	//		assert.NoError(t, err)
	//		assert.Equal(t, out.Data, d)
	//		assert.True(t, strings.HasSuffix(out.Version, "latest"))
	//		assert.Equal(t, len(out.Versions), 2)
	//	})
	//})
}
