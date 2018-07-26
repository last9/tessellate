package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"strings"
	"time"

	"fmt"

	"os"

	"github.com/meson10/highbrow"
	"github.com/meson10/pester"
	"github.com/pkg/errors"
	"gitlab.com/tsocial/sre/tessellate/runner"
	"gitlab.com/tsocial/sre/tessellate/storage/consul"
	"gitlab.com/tsocial/sre/tessellate/storage/types"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Version of the runner.
const Version = "0.0.1"

var (
	jobID       = kingpin.Flag("job", "Job ID").Short('j').String()
	workspaceID = kingpin.Flag("workspace", "Workspace ID").Short('w').String()
	layoutID    = kingpin.Flag("layout", "Layout ID").Short('l').String()
	consulIP    = kingpin.Flag("consul-host", "Consul IP").Short('c').String()
)

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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	kingpin.Version(Version)
	kingpin.Parse()

	// Initialize Storage engine
	store := consul.MakeConsulStore(*consulIP)
	store.Setup()

	status := 0

	defer func() {
		key := fmt.Sprintf("%v-%v", *workspaceID, *layoutID)

		if err := highbrow.Try(3, func() error {
			return store.Unlock(key, *jobID)
		}); err != nil {
			log.Printf("error while unlocking key: %s, err: %+v", key, err)
		}

		os.Exit(status)
	}()

	// Create Job Struct to Load Job into.
	j := types.Job{Id: *jobID, LayoutId: *layoutID}
	t := types.MakeTree(*workspaceID)
	if err := store.GetVersion(&j, t, *jobID); err != nil {
		log.Println(err)
		status = 127
		return
	}

	// Make Layout tree.

	// Get Layout
	l := types.Layout{Id: j.LayoutId}
	if err := store.GetVersion(&l, t, j.LayoutVersion); err != nil {
		log.Println(err)
		status = 127
		return
	}

	// Get Workspace Vars
	var wv types.Vars
	if err := store.Get(&wv, t); err != nil {
		log.Println(err)
		if !strings.Contains(err.Error(), "Missing") {
			status = 127
			return
		}
	}

	if err := padLayoutWithProvider(l.Plan, wv); err != nil {
		status = 127
		return
	}

	// Get Vars
	var v types.Vars
	t2 := types.MakeTree(*workspaceID, j.LayoutId)
	if err := store.GetVersion(&v, t2, j.VarsVersion); err != nil {
		log.Println(err)
		if !strings.Contains(err.Error(), "Missing") {
			status = 127
			return
		}
	}

	remotePath := path.Join("state", *workspaceID, j.LayoutId)
	startState, _ := store.GetKey(remotePath)

	op := j.Op
	if j.Dry {
		op = runner.PlanOp
	}

	cmd := runner.Cmd{}
	cmd.SetOp(op)
	cmd.SetRemotePath(remotePath)
	cmd.SetRemote(*consulIP)
	cmd.SetDir("/tmp/test_runner")
	cmd.SetLayout(l.Plan)
	cmd.SetVars(v)
	cmd.SetLogPrefix(j.Id)

	var w types.Watch
	if err := store.GetVersion(&w, t2, "latest"); err != nil {
		log.Println(err)
	}

	url := w.SuccessURL

	if err := cmd.Run(); err != nil {
		status = 127
		log.Println(err)
		url = w.FailureURL
	}

	if url != "" {
		endState, _ := store.GetKey(remotePath)
		body := struct {
			OldState interface{} `json:"old_state"`
			NewState interface{} `json:"new_state"`
		}{}

		if err := json.Unmarshal(startState, &body.OldState); err != nil {
			errors.Wrap(err, "Cannot unmarshal")
		}

		if err := json.Unmarshal(endState, &body.NewState); err != nil {
			errors.Wrap(err, "Cannot unmarshal")
		}

		bfinal, _ := json.Marshal(body)
		if req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bfinal)); err != nil {
			log.Println(err)
		} else {
			makeCall(req)
		}
	}
}
