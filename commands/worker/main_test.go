package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"

	"net/http"

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

		collector := map[string][]*watchPacket{}

		s, err := hookServer(func(uv *url.URL, p *watchPacket) {
			u := uv.String()
			if _, ok := collector[u]; !ok {
				collector[u] = []*watchPacket{}
			}
			collector[u] = append(collector[u], p)
		})
		if err != nil {
			t.Fatal(err)
		}

		defer s.Close()

		defaultWatch := "/default-watch"
		h, err := url.Parse(s.URL + defaultWatch)
		if err != nil {
			t.Fatal(err)
		}

		tree := types.MakeTree(wID, lID)
		if err := addWatch(tree, s.URL); err != nil {
			t.Fatal(err)
		}

		t.Run("without default watch", func(t *testing.T) {
			x := mainRunner(store, in, nil)
			// expect that 2 handlers are called.
			assert.Equal(t, 0, x)
			assert.Equal(t, 0, len(collector[defaultWatch]))
			assert.Equal(t, 1, len(collector))
		})

		t.Run("with default watch", func(t *testing.T) {
			x := mainRunner(store, in, h)
			// expect that 2 handlers are called.
			assert.Equal(t, 0, x)
			assert.Equal(t, 1, len(collector[defaultWatch]))
			assert.Equal(t, 2, len(collector))
		})
	})

	t.Run("Should fail", func(t *testing.T) {
		layoutSave("../../runner/testdata/faulty.tf.json")
		in := &input{
			jobID:       jID,
			workspaceID: wID,
			layoutID:    lID,
			tmpDir:      "failed-run",
		}

		x := mainRunner(store, in, nil)
		assert.Equal(t, 127, x)
	})
}

// Starts a new server for test purposes
func watchMux(f func(*url.URL, *watchPacket)) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	mux.Handle("/user-watch", watchHandler(f))
	mux.Handle("/default-watch", watchHandler(f))
	return mux, nil
}

func watchHandler(f func(url *url.URL, p *watchPacket)) http.HandlerFunc {
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

		f(r.URL, &s)
		w.WriteHeader(http.StatusOK)
	}
}

func hookServer(f func(*url.URL, *watchPacket)) (*httptest.Server, error) {
	var err error
	m, err := watchMux(f)
	if err != nil {
		return nil, err
	}

	s := httptest.NewServer(m)
	return s, nil
}

func addWatch(tree *types.Tree, url string) error {
	uw := types.Watch{SuccessURL: url + "/user-watch"}
	if err := store.Save(&uw, tree); err != nil {
		return err
	}
	return nil
}
