package main

import (
	"bytes"
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

	cmd := runner.Cmd{}
	cmd.SetOp(j.Op)
	cmd.SetRemotePath(path.Join("state", *workspaceID, j.LayoutId))
	cmd.SetRemote(*consulIP)
	cmd.SetDir("/tmp/test_runner")
	cmd.SetLayout(l.Plan)
	cmd.SetVars(v)
	cmd.SetLogPrefix(j.Id)

	var w types.Watch
	if err := store.GetVersion(&w, t2, "latest"); err != nil {
		log.Println(err)
	}

	var req *http.Request
	var err2 error

	status := 0

	if err := cmd.Run(); err != nil {
		status = 127
		log.Println(err)
		if w.FailureURL != "" {
			req, err2 = http.NewRequest(http.MethodPost, w.FailureURL, bytes.NewBuffer(nil))
		}
	} else {
		if w.SuccessURL != "" {
			req, err2 = http.NewRequest(http.MethodPost, w.SuccessURL, bytes.NewBuffer(nil))
		}
	}

	if err2 != nil {
		log.Println(err2)
	}

	if req != nil {
		makeCall(req)
	}

	os.Exit(status)
}
