package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"

	"net/http"

	"fmt"
	"net/http/httptest"
	"net/url"

	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/server"
	"github.com/tsocial/tessellate/storage"
	"github.com/tsocial/tessellate/storage/types"
)

var store storage.Storer

func TestMainRunner(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	wID := "w123"
	lID := "l123"

	// Tree for workspace ID.
	tree := types.MakeTree(wID)

	func() {
		// Create a new types.Workspace instance to be returned.
		workspace := types.Workspace(wID)
		err := store.Save(&workspace, tree)
		assert.Nil(t, err)
	}()

	layoutSave := func(path string) {
		plan := map[string]json.RawMessage{}
		lBytes, err := ioutil.ReadFile(path)
		assert.Nil(t, err)

		plan["sleep"] = lBytes

		// Create layout instance to be saved for given ID and plan.
		layout := types.Layout{Id: lID, Plan: plan}

		// Save the layout.
		err = store.Save(&layout, tree)
		assert.Nil(t, err)
	}

	jID := func() string {
		j := types.Job{
			LayoutId:      lID,
			LayoutVersion: "latest",
			Op:            int32(server.Operation_APPLY),
		}

		// Save this job in workspace tree.
		err := store.Save(&j, tree)
		assert.Nil(t, err)

		return j.Id
	}()

	t.Run("Should run successfully", func(t *testing.T) {
		layoutSave("../../runner/testdata/sleep.tf.json")
		in := &input{
			jobID:       jID,
			workspaceID: wID,
			layoutID:    lID,
			tmpDir:      "success-run",
		}

		s, err := attachWatch(wID, lID)
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()

		x := mainRunner(store, in)
		assert.Equal(t, 0, x)
	})

	t.Run("Should fail", func(t *testing.T) {
		layoutSave("../../runner/testdata/faulty.tf.json")
		in := &input{
			jobID:       jID,
			workspaceID: wID,
			layoutID:    lID,
			tmpDir:      "failed-run",
		}

		x := mainRunner(store, in)
		assert.Equal(t, 127, x)
	})
}

// Starts a new server for test purposes
func watchMux(f func(p *watchPacket)) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	mux.Handle("/user-watch", watchHandler(f))
	mux.Handle("/default-watch", watchHandler(f))
	return mux, nil
}

func watchHandler(f func(p *watchPacket)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := watchPacket{}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(b, &s); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		f(&s)
		w.WriteHeader(http.StatusOK)
	}
}

func attachWatch(workspaceID, layoutID string) (*httptest.Server, error) {
	var err error
	m, err := watchMux(func(p *watchPacket) {
		fmt.Println(p)
	})
	if err != nil {
		return nil, err
	}

	s := httptest.NewServer(m)
	tree := types.MakeTree(workspaceID, layoutID)

	uw := types.Watch{SuccessURL: s.URL + "/user-watch"}
	if err := store.Save(&uw, tree); err != nil {
		return nil, err
	}

	*defaultHook, err = url.Parse(s.URL + "/default-watch")
	if err != nil {
		return nil, err
	}
	return s, nil
}
