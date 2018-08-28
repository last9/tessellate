package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"os"

	"github.com/meson10/highbrow"
	"github.com/meson10/pester"
	"github.com/pkg/errors"
	"github.com/tsocial/tessellate/runner"
	"github.com/tsocial/tessellate/storage"
	"github.com/tsocial/tessellate/storage/consul"
	"github.com/tsocial/tessellate/storage/types"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Version of the runner.
const Version = "0.0.1"

var (
	jobID       = kingpin.Flag("job", "Job ID").Short('j').String()
	workspaceID = kingpin.Flag("workspace", "Workspace ID").Short('w').String()
	layoutID    = kingpin.Flag("layout", "Layout ID").Short('l').String()
	consulIP    = kingpin.Flag("consul-host", "Consul IP").Short('c').String()
	tmpDir      = kingpin.Flag("tmp-dir", "Temporary Dir").Short('d').Default("test-runner").String()
)

type input struct {
	jobID       string
	workspaceID string
	layoutID    string
	tmpDir      string
}

// Make a HTTP Call to the callbacks specified.
// Does an internal retries in case of connection failures.
func makeCall(req *http.Request) error {
	log.Println("Making request", req)
	client := pester.New()
	client.Concurrency = 3
	client.MaxRetries = 3
	client.Backoff = pester.ExponentialBackoff
	client.KeepLog = true
	client.Timeout = time.Duration(5 * time.Second)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Println("Response from callback", string(b))
	return nil
}

// Exit Routine that is responsible for cleaning the Lock, only in case of success.
// Since failure ends up retrying a job on the Job scheduler, Lock cleanup should be a
// responsibility of dispatcher in case of failures.
// Cleanup lies outside the scope of this worker.
func exit(store storage.Storer, status int) {
	if status == 0 {
	}
	os.Exit(status)
}

// getJob details.
func getJob(store storage.Storer, in *input) (*types.Job, error) {
	j := types.Job{Id: in.jobID, LayoutId: in.layoutID}
	t := types.MakeTree(in.workspaceID)
	if err := store.GetVersion(&j, t, in.jobID); err != nil {
		return nil, errors.Wrap(err, "Cannot Load job")
	}
	return &j, nil
}

// getLayout details
func getLayout(store storage.Storer, j *types.Job, in *input) (*types.Layout, error) {
	// Get Layout
	l := types.Layout{Id: in.layoutID}
	t := types.MakeTree(in.workspaceID)
	if err := store.GetVersion(&l, t, j.LayoutVersion); err != nil {
		return nil, errors.Wrap(err, "Cannot load Layout")
	}
	return &l, nil
}

// getWorkspaceVars, but don't raise error if not found.
// Please note, this always returns the latest value of the variables since they are set
// by workspace administrators.
func getWorkspaceVars(store storage.Storer, in *input) (*types.Vars, error) {
	// Get Workspace Vars
	var wv types.Vars
	t := types.MakeTree(in.workspaceID)
	if err := store.Get(&wv, t); err != nil {
		if !strings.Contains(err.Error(), "Missing") {
			return nil, errors.Wrap(err, "Cannot find workspaace vars")
		}
	}

	return &wv, nil
}

// getLayoutWatch looks for any active Watch on this Layout. Doesnt return error if not found.
func getLayoutWatch(store storage.Storer, in *input) (types.Watch, error) {
	w := types.Watch{}
	t := types.MakeTree(in.workspaceID, in.layoutID)
	if err := store.GetVersion(&w, t, "latest"); err != nil {
		if !strings.Contains(err.Error(), "Missing") {
			return w, errors.Wrap(err, "Cannot find Layout watch")
		}
	}
	return w, nil
}

// getJobVars gets the variables that were specified at the time of Job creation.
// Unlike workspace vars Version is important here to fetch point-in-time value of the vars,
// since they would have moved on by the time job was actually run.
func getJobVars(store storage.Storer, j *types.Job, in *input) (*types.Vars, error) {
	// Get Vars
	var v types.Vars
	t2 := types.MakeTree(*workspaceID, j.LayoutId)
	if err := store.GetVersion(&v, t2, j.VarsVersion); err != nil {
		log.Println(err)
		if !strings.Contains(err.Error(), "Missing") {
			return nil, errors.Wrap(err, "Cannot find job vars")
		}
	}

	return &v, nil
}

func remotePath(in *input) string {
	return path.Join("state", in.workspaceID, in.layoutID)
}

func getCmd(store storage.Storer, in *input) (*runner.Cmd, error) {
	j, err := getJob(store, in)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get Job")
	}

	l, err := getLayout(store, j, in)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get Layout")
	}

	wv, err := getWorkspaceVars(store, in)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get workspace vars")
	}

	if err := padLayoutWithProvider(l.Plan, wv); err != nil {
		return nil, errors.Wrap(err, "Cannot pad layout")
	}

	v, err := getJobVars(store, j, in)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot get job vars")
	}

	op := j.Op
	if j.Dry {
		op = runner.PlanOp
	}

	cmd := runner.Cmd{}
	cmd.SetOp(op)
	cmd.SetRemotePath(remotePath(in))
	cmd.SetRemote(*consulIP)
	cmd.SetDir(path.Join("/tmp", in.tmpDir))
	cmd.SetLayout(l.Plan)
	cmd.SetVars(*v)
	cmd.SetLogPrefix(j.Id)
	return &cmd, nil
}

// Engine tries to accept a storage and input and run the Command.
func engine(store storage.Storer, in *input) (string, error) {
	cmd, err := getCmd(store, in)
	if err != nil {
		return "", errors.Wrap(err, "Cannot get cmd")
	}

	w, err := getLayoutWatch(store, in)
	if err != nil {
		return w.SuccessURL, errors.Wrap(err, "Cannot get layout watch")
	}

	url := w.SuccessURL
	if err := cmd.Run(); err != nil {
		return w.FailureURL, errors.Wrap(err, "Exited with failure")
	}

	return url, errors.Wrap(err, "Error executing Cmd")
}

// MainRunner takes input parametes and does the rest.
// Invokes the engine, and also makes callback to any watches that may be available
// on the layout.
func mainRunner(store storage.Storer, in *input) int {
	status := 0

	if err := func() error {
		startState, _ := store.GetKey(remotePath(in))
		url, err := engine(store, in)
		if err != nil {
			return errors.Wrap(err, "Cannot execute Engine")
		}

		if url == "" {
			return nil
		}

		endState, _ := store.GetKey(remotePath(in))
		body := struct {
			OldState interface{} `json:"old_state"`
			NewState interface{} `json:"new_state"`
		}{}

		if err := json.Unmarshal(startState, &body.OldState); err != nil {
			log.Println(err)
			return nil
		}

		if err := json.Unmarshal(endState, &body.NewState); err != nil {
			log.Println(err)
			return nil
		}

		bfinal, _ := json.Marshal(body)
		if req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bfinal)); err != nil {
			log.Println(err)
		} else {
			makeCall(req)
		}

		return nil
	}(); err != nil {
		fmt.Printf("%+v\n", err)
		status = 127
	} else {
		// UnLock Lock for workspace and layout.
		key := fmt.Sprintf("%v-%v", *workspaceID, *layoutID)
		highbrow.Try(5, func() error {
			return store.Unlock(key)
		})
	}

	return status
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	kingpin.Version(Version)
	kingpin.Parse()

	// Initialize Storage engine
	store := consul.MakeConsulStore(*consulIP)
	store.Setup()

	in := &input{
		jobID:       *jobID,
		workspaceID: *workspaceID,
		layoutID:    *layoutID,
		tmpDir:      *tmpDir,
	}

	os.Exit(mainRunner(store, in))
}
