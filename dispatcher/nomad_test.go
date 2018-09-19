// +build nomad

package dispatcher

import (
	"os"
	"testing"

	"github.com/hashicorp/nomad/api"
	"github.com/stretchr/testify/assert"
	"github.com/tsocial/tessellate/storage/types"
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
	var runningJob *api.Job

	t.Run("Should contain workspace and layout id in the nomad job.", func(t *testing.T) {
		workspaceID := "workspace"

		nomadClient, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			t.Fatal(err)
		}

		jobName := workspaceID + "-" + job.LayoutId + "-" + job.Id

		if runningJob, _, err = nomadClient.Jobs().Info(jobName, nil); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, jobName, *runningJob.Name)

	})

	t.Run("Only job ID should be appended in the task config, not workspace and layout ID", func(t *testing.T) {
		tasks := runningJob.TaskGroups[0]
		for _, val := range tasks.Tasks {
			jobID := val.Config["entrypoint"].([]interface{})[2]

			if jobID != job.Id {
				t.Fatal("Job ID expected %v, Got %v", job.Id, jobID)
			}
		}
	})
}
