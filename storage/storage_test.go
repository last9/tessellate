package storage

import (
	"log"
	"testing"

	"math/rand"
	"time"

	"strings"

	"os"

	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/storage/consul"
	"github.com/tsocial/tessellate/storage/types"
	"github.com/tsocial/tessellate/utils"
)

var store Storer

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

	os.Exit(func() int {
		defer deleteTree(store.GetClient())

		y := m.Run()
		return y
	}())
}

func TestStorer(t *testing.T) {
	t.Run("Storage tests", func(t *testing.T) {
		tree := &types.Tree{Name: "store_test", TreeType: "testing"}

		workspace := types.Workspace(fmt.Sprintf("alibaba-%s", utils.RandString(8)))

		t.Run("Workspace does not exist", func(t *testing.T) {
			if err := store.Get(&workspace, tree); err == nil {
				t.Fatal("Should have failed with an Error")
			}
		})

		t.Run("Get a Workspace after creation", func(t *testing.T) {
			err := store.Save(&workspace, tree)
			if err != nil {
				t.Fatal(err)
			}

			if err := store.Get(&workspace, tree); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("Resaving a Workspace doesn't raise an Error", func(t *testing.T) {
			if err := store.Save(&workspace, tree); err != nil {
				t.Fatal(err)
			}

			if err := store.Get(&workspace, tree); err != nil {
				t.Fatal(err)
			}

			v, err := store.GetVersions(&workspace, tree)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, 3, len(v))
			assert.Contains(t, strings.Join(v, ""), "latest")
		})
	})
}
