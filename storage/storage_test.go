package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/storage/types"
	"github.com/tsocial/tessellate/utils"
)

var store Storer

func TestStorer(t *testing.T) {
	t.Run("Lock tests", func(t *testing.T) {
		t.Run("Lock a Key", func(t *testing.T) {
			if err := store.Lock("key3", "c1"); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("Un-Idempotent Lock a Key", func(t *testing.T) {
			if err := store.Lock("key3", "c12"); err == nil {
				t.Fatal("Should have raised a key")
			}
		})

		t.Run("Release a Key", func(t *testing.T) {
			if err := store.Unlock("key3"); err != nil {
				t.Fatal(err)
			}
		})

		t.Run("Idempotent Release a Key", func(t *testing.T) {
			if err := store.Unlock("key3"); err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("Storage tests", func(t *testing.T) {
		tree := &types.Tree{Name: "store_test", TreeType: "testing"}

		wid := fmt.Sprintf("alibaba-%s", utils.RandString(8))
		workspace := types.Workspace(wid)

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

		t.Run("Re-saving a Workspace doesn't raise an Error", func(t *testing.T) {
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

		t.Run("Save Layout", func(t *testing.T) {
			tree := types.MakeTree(wid)
			l := types.Layout{
				Id:   "test-hello",
				Plan: map[string]json.RawMessage{},
			}

			if err := store.Save(&l, tree); err != nil {
				t.Fatal(err)
			}

			ltree := types.MakeTree(wid, "test-hello")
			v := types.Vars(map[string]interface{}{})
			if err := store.Save(&v, ltree); err != nil {
				t.Fatal(err)
			}

			x, err := store.GetVersions(&l, tree)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, 2, len(x))
		})

		t.Run("Get absent Key", func(t *testing.T) {
			d, err := store.GetKey("hello/world")
			assert.Nil(t, err)
			assert.Equal(t, []byte{}, d)
		})

		t.Run("Get Valid Key", func(t *testing.T) {
			key := fmt.Sprintf("workspaces/%v/latest", wid)
			d, err := store.GetKey(key)
			assert.Nil(t, err)
			assert.NotEqual(t, []byte{}, d)
		})

		t.Run("Get Keys", func(t *testing.T) {
			prefix := types.WORKSPACE + "/"
			separator := "/"

			keys, err := store.GetKeys(prefix, separator)
			assert.Nil(t, err)

			log.Println(keys)
			for _, k := range keys {
				splits := strings.Split(k, "/")
				assert.Equal(t, 3, len(splits))
			}
		})

		t.Run("Save and Get Key", func(t *testing.T) {
			key := uuid.NewV4().String()
			val := uuid.NewV4().String()

			assert.Nil(t, store.SaveKey(key, []byte(val)))

			got, err := store.GetKey(key)
			assert.Nil(t, err)
			assert.Equal(t, val, string(got))
		})
	})
}
