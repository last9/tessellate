package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/meson10/pester"
	"github.com/tsocial/tessellate/runner"
	"github.com/tsocial/tessellate/storage/consul"
	"github.com/tsocial/tessellate/storage/types"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Version of the runner.
const Version = "0.0.1"

var (
	jobID       = kingpin.Flag("job", "Job ID").Short('j').String()
	workspaceID = kingpin.Flag("workspace", "Workspace ID").Short('w').String()
	consulIP    = kingpin.Flag("consul-host", "Consul IP").Short('c').String()
)

func makeCall(req *http.Request) error {
	log.Println("Making rqeuest", req)
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

	// Inititalize Storage engine
	store := consul.MakeConsulStore(*consulIP)
	store.Setup()

	// Create Job Struct to Load Job into.
	j := types.Job{Id: *jobID}
	t := types.MakeTree(*workspaceID)
	if err := store.Get(&j, t); err != nil {
		log.Println(err)
		os.Exit(127)
	}

	// Make Layout tree.
	t2 := types.MakeTree(*workspaceID, j.LayoutId)
	remotePath := path.Join("state", *workspaceID, j.LayoutId)

	// Get Layout
	l := types.Layout{Id: j.LayoutId}
	if err := store.GetVersion(&l, t, j.LayoutVersion); err != nil {
		log.Println(err)
		os.Exit(127)
	}

	// Get Vars
	var v types.Vars
	if err := store.GetVersion(&v, t2, j.VarsVersion); err != nil {
		log.Println(err)
		if !strings.Contains(err.Error(), "Missing") {
			os.Exit(127)
		}
	}

	startState, _ := store.GetKey(remotePath)

	cmd := runner.Cmd{}
	cmd.SetOp(j.Op)
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

	status := 0
	url := w.SuccessURL

	if err := cmd.Run(); err != nil {
		status = 127
		log.Println(err)
		url = w.FailureURL
	}

	if url != "" {
		endState, _ := store.GetKey(remotePath)
		body := struct {
			OldState []byte `json:"old_state"`
			NewState []byte `json:"new_state"`
		}{OldState: startState, NewState: endState}

		b, _ := json.Marshal(body)
		if req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b)); err != nil {
			log.Println(err)
		} else {
			makeCall(req)
		}
	}

	os.Exit(status)
}
