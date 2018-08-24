package dispatcher

import (
	"log"

	"github.com/flosch/pongo2"
	"github.com/hashicorp/nomad/api"
	"github.com/tsocial/tessellate/storage/types"
	"github.com/tsocial/tessellate/tmpl"
)

const defaultAttempts = 3

type client struct {
	cfg NomadConfig
}

type NomadConfig struct {
	Address    string
	Username   string
	Password   string
	Datacenter string
	Image      string
	CPU        string
	Memory     string
	ConsulAddr string
}

func NewNomadClient(cfg NomadConfig) *client {
	return &client{cfg}
}

func (c *client) Dispatch(w string, j *types.Job) error {
	// Create a nomad job using go template
	var tmplStr = `
job "{{ job_id }}" {
  datacenters = ["{{ datacenter }}"]
  type        = "batch"

  group "{{ job_id }}" {
    count = 1

	restart {
      attempts = {{ attempts }}
    }

    task "apply_job" {
      driver = "docker"

      config {
        image = "{{ image }}"
        entrypoint = ["./tsl8", "-j", "{{ job_id }}", "-w", "{{ workspace_id }}", "-l", "{{ layout_id }}", "--consul-host", "{{ consul_addr }}"]
      }

      resources {
        cpu    = {{ cpu }}
        memory = {{ memory }}
      }
    }
  }
}
`
	cfg := pongo2.Context{
		"job_id":       j.Id,
		"workspace_id": w,
		"layout_id":    j.LayoutId,
		"datacenter":   c.cfg.Datacenter,
		"image":        c.cfg.Image,
		"cpu":          c.cfg.CPU,
		"memory":       c.cfg.Memory,
		"consul_addr":  c.cfg.ConsulAddr,
		"attempts":     defaultAttempts,
	}

	if j.Dry {
		cfg["attempts"] = 0
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

	log.Printf("successfully dispatched the job: %+v", resp)
	return nil
}
