// +build integration

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
	j, err := Get().Dispatch("workspace", job)
	assert.Nil(t, err)
	assert.NotNil(t, j)
}

func TestDispatched_Job(t *testing.T) {
	job := &types.Job{Id: "job", LayoutId: "layout"}
	var runningJob *api.Job

	t.Run("Should contain workspace and layout id in the nomad job.", func(t *testing.T) {
		workspaceID := "workspace"

		nomadClient, err := api.NewClient(api.DefaultConfig())
		assert.Nil(t, err)

		jobName := workspaceID + "-" + job.LayoutId + "-" + job.Id

		runningJob, _, err = nomadClient.Jobs().Info(jobName, nil)
		assert.Nil(t, err)

		assert.Equal(t, jobName, *runningJob.Name)

	})
	t.Run("Only job ID should be appended in the task config, not workspace and layout ID", func(t *testing.T) {
		tasks := runningJob.TaskGroups[0]
		for _, val := range tasks.Tasks {
			jobID := val.Config["entrypoint"].([]interface{})[2]
            assert.Equal(t, jobID, job.Id)
		}
	})
}
