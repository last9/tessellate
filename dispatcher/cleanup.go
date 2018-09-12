package dispatcher

import (
	"fmt"
	"log"

	"github.com/flosch/pongo2"
	"github.com/hashicorp/nomad/api"
	"github.com/tsocial/tessellate/tmpl"
)

func (c *client) GetOrSetCleanup(s string) error {
	nConfig := api.DefaultConfig()
	nConfig.Address = c.cfg.Address

	if c.cfg.Username != "" {
		nConfig.HttpAuth = &api.HttpBasicAuth{
			Username: c.cfg.Username,
			Password: c.cfg.Password,
		}
	}

	cl, err := api.NewClient(nConfig)
	if err != nil {
		log.Printf("error while creating nomad client: %+v", err)
		return err
	}

	job, _, err := cl.Jobs().Info(s, nil)
	// Replace the Job if not found OR existing Job was dead.
	if job == nil || *job.Status == "dead" {
		if err := c.startCleanupJob(s); err != nil {
			return err
		}
	}

	return nil
}

func cleanupCmd(nomadHost, consulHost string) string {
	return fmt.Sprintf(`curl -XGET http://%v/v1/kv/lock/?keys | jq -c '.[]' | tr -d '\"' | xargs -i bash -c 'echo {} > /tmp/job_id && curl -XGET %v/v1/kv/{}' | jq -c '.[] | (.Value)' | tr -d '\"' | base64 --decode | xargs -i curl -XGET %v/v1/job/{} | jq 'select( .Status == \"dead\")' && curl -XDELETE http://%v/v1/kv/$(cat /tmp/job_id)`,
		consulHost, consulHost, nomadHost, consulHost)
}

func (c *client) startCleanupJob(jobID string) error {
	// Create a nomad job using go template
	var tmplStr = fmt.Sprintf(`
job "{{ job_id }}" {
  datacenters = ["{{ datacenter }}"]
  type        = "batch"

  periodic {
    cron             = "*/2 * * * * *"
    prohibit_overlap = true
  }

  group "{{ job_id }}" {
    count = 1

    task "cleanup_job" {
      driver = "raw_exec"

		config {
			command = "bash"
			args = ["-exc", "%v"]
      }

      resources {
        cpu    = {{ cpu }}
        memory = {{ memory }}
      }
    }
  }
}
`, cleanupCmd(c.cfg.Address, c.cfg.ConsulAddr))

	cfg := pongo2.Context{
		"job_id":     jobID,
		"datacenter": c.cfg.Datacenter,
		"cpu":        c.cfg.CPU,
		"memory":     c.cfg.Memory,
	}

	nomadJob, err := tmpl.Parse(tmplStr, cfg)
	if err != nil {
		log.Printf("error while job parsing: %+v", err)
		return err
	}

	log.Println(nomadJob)

	nConfig := api.DefaultConfig()
	nConfig.Address = c.cfg.Address

	if c.cfg.Username != "" {
		nConfig.HttpAuth = &api.HttpBasicAuth{
			Username: c.cfg.Username,
			Password: c.cfg.Password,
		}
	}

	cl, err := api.NewClient(nConfig)
	if err != nil {
		log.Printf("error while creating nomad client: %+v", err)
		return err
	}

	jobs := cl.Jobs()
	job, err := jobs.ParseHCL(nomadJob, true)
	if err != nil {
		log.Printf("error while parsing job hcl: %+v", err)
		return err
	}

	resp, _, err := jobs.Register(job, nil)
	if err != nil {
		log.Printf("error while registering nomad job: %+v", err)
		return err
	}

	log.Printf("successfully started the job: %+v", resp)
	return nil
}
