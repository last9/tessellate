package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/server"
	"github.com/tsocial/tessellate/storage/consul"
	"github.com/tsocial/tessellate/storage/types"
)

func TestMainRunner(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	wID := "w123"
	lID := "l123"

	log.Printf("waiting for consul")
	time.Sleep(30 * time.Second)
	store := consul.MakeConsulStore(os.Getenv("CONSUL_ADDR"))
	store.Setup()

	defer func() {
		store.GetClient().KV().DeleteTree(wID+"/", &api.WriteOptions{})
	}()

	// Tree for workspace ID.
	tree := types.MakeTree(wID)

	func() {
		// Create a new types.Workspace instance to be returned.
		workspace := types.Workspace(wID)
		if err := store.Save(&workspace, tree); err != nil {
			t.Fatal(err)
		}
	}()

	layoutSave := func(path string) {
		plan := map[string]json.RawMessage{}
		lBytes, err := ioutil.ReadFile(path)
		if err != nil {
			t.Error(err)
		}

		plan["sleep"] = lBytes
		// Create layout instance to be saved for given ID and plan.
		layout := types.Layout{Id: lID, Plan: plan}

		// Save the layout.
		if err := store.Save(&layout, tree); err != nil {
			t.Fatal(err)
		}
	}

	jID := func() string {
		j := types.Job{
			LayoutId:      lID,
			LayoutVersion: "latest",
			Op:            int32(server.Operation_APPLY),
		}

		// Save this job in workspace tree.
		if err := store.Save(&j, tree); err != nil {
			t.Fatal(err)
		}

		return j.Id
	}()

	t.Run("Should run successfully", func(t *testing.T) {
		layoutSave("../../runner/testdata/sleep.tf.json")
		in := &input{
			jobID:       jID,
			workspaceID: wID,
			layoutID:    lID,
		}

		x := mainRunner(store, in)
		assert.Equal(t, 0, x)
	})

	t.Run("Should fail", func(t *testing.T) {
		layoutSave("../../runner/testdata/faulty.tf.json")
		in := &input{
			jobID:       jID,
			workspaceID: wID,
			layoutID:    lID,
		}

		x := mainRunner(store, in)
		assert.Equal(t, 127, x)
	})
}
