package dispatcher

import (
	"fmt"
	"log"

	"github.com/flosch/pongo2"
	"github.com/hashicorp/nomad/api"
	"github.com/tsocial/tessellate/tmpl"
)

const CLEANUP_PYTHON = `
import urllib2
import json
import time

while True:
    try:
        keys = json.loads(urllib2.urlopen("http://%v/v1/kv/lock/?keys").read().decode('utf-8'))[1:]
    except urllib2.HTTPError as e:
        print("No lock found")

    if keys and len(keys) > 0:
        print("Keys are : " + ', '.join(keys))
        for key in keys:
            response = json.loads(urllib2.urlopen("http://%v/v1/kv/" + key).read().decode('utf-8'))[0]
            k = response['Key'].split('/')[1]
            v = response['Value'].decode('base64')
            print("Fetching Nomad Job details for Nomad Job Name: " + k + "-" + v)
            status = None
            try:
                status = json.loads(urllib2.urlopen("%v/v1/job/" + k + "-" + v).read().decode('utf-8'))['Status']
            except urllib2.HTTPError:
                print("Job not found")
            if status and status == 'dead':
                print("Deleting lock from Consul for dead Nomad Job: " + key)
                opener = urllib2.build_opener(urllib2.HTTPHandler)
                request = urllib2.Request("http://%v/v1/kv/" + key, data='')
                request.get_method = lambda: 'DELETE'
                opener.open(request)
    time.sleep(5)
`

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
	return fmt.Sprintf(CLEANUP_PYTHON,
		consulHost, consulHost, nomadHost, consulHost)
}

func (c *client) startCleanupJob(jobID string) error {
	// Create a nomad job using go template
	var tmplStr = fmt.Sprintf(`
job "{{ job_id }}" {
  datacenters = ["{{ datacenter }}"]
  type        = "service"

  group "{{ job_id }}" {
    count = 1

    task "cleanup_job" {
      driver = "raw_exec"

      template {
		data = <<EOH
%v
EOH
		destination = "/tmp/unlock.py"
		perms       = 755
	  }

      config {
		command = "python"
		args = ["tmp/unlock.py"]
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
