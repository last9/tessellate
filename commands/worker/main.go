package main

import (
	"log"
	"os"

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
	if err := store.GetVersion(&l, t2, j.LayoutVersion); err != nil {
		log.Println(err)
		os.Exit(127)
	}

	// Get Vars
	var v types.Vars
	if err := store.GetVersion(&v, t2, j.VarsVersion); err != nil {
		log.Println(err)
		os.Exit(127)
	}

	cmd := runner.Cmd{}
	cmd.SetOp(j.Op)
	cmd.SetRemote(*consulIP)
	cmd.SetDir("/tmp/test_runner")
	cmd.SetLayout(l.Plan)
	cmd.SetVars(v)
	cmd.SetLogPrefix(j.Id)

	if err := cmd.Run();err != nil {
		log.Println(err)
		os.Exit(127)
	}
}
