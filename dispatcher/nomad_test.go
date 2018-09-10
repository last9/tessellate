// +build nomad

package dispatcher

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/storage/types"
	"github.com/hashicorp/nomad/api"
)

func TestMain(m *testing.M) {

	cl := NewNomadClient(NomadConfig{
		Address:    "http://127.0.0.1:4646",
		Datacenter: "dc1",
		Image:      "redis",
		CPU:        "32",
		Memory:     "32",
	})

	Set(cl)
	os.Exit(m.Run())
}

func TestClient_Dispatch(t *testing.T) {
	job := &types.Job{Id: "job", LayoutId: "layout"}
	assert.Nil(t, Get().Dispatch("workspace", job))
}

func TestDispatched_Job(t *testing.T) {
	job := &types.Job{Id: "job", LayoutId: "layout"}
	workspaceID := "workspace"

	nomadClient , err := api.NewClient(api.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	jobID := workspaceID + "-" + job.LayoutId + "-" + job.Id

	runningJob,_, err := nomadClient.Jobs().Info(jobID, nil)
	if err!= nil {
		t.Fatal(err)
	}

	assert.Equal(t, jobID, *runningJob.Name)

}