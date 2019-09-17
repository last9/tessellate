package dispatcher

import (
	"fmt"
	"log"
	"os"

	"github.com/flosch/pongo2"
	"github.com/hashicorp/nomad/api"
	"github.com/tsocial/tessellate/tmpl"
)

const CLEANUP_PYTHON = `
import urllib2
import json
import httplib
import ssl
import socket
import logging
import time

DEFAULT_HTTP_TIMEOUT = 10 #seconds

# http://code.activestate.com/recipes/577548-https-httplib-client-connection-with-certificate-v/
# http://stackoverflow.com/questions/1875052/using-paired-certificates-with-urllib2

class HTTPSClientAuthHandler(urllib2.HTTPSHandler):
    '''
    Allows sending a client certificate with the HTTPS connection.
    This version also validates the peer (server) certificate since, well...
    WTF IS THE POINT OF SSL IF YOU DON"T AUTHENTICATE THE PERSON YOU"RE TALKING TO!??!
    '''
    def __init__(self, key=None, cert=None, ca_certs=None, ssl_version=None, ciphers=None):
        urllib2.HTTPSHandler.__init__(self)
        self.key = key
        self.cert = cert
        self.ca_certs = ca_certs
        self.ssl_version = ssl_version
        self.ciphers = ciphers

    def https_open(self, req):
        # Rather than pass in a reference to a connection class, we pass in
        # a reference to a function which, for all intents and purposes,
        # will behave as a constructor
        return self.do_open(self.get_connection, req)

    def get_connection(self, host, timeout=DEFAULT_HTTP_TIMEOUT):
        return HTTPSConnection( host,
                key_file = self.key,
                cert_file = self.cert,
                timeout = timeout,
                ciphers = self.ciphers,
                ca_certs = self.ca_certs,
                ssl_version = self.ssl_version )


class HTTPSConnection(httplib.HTTPSConnection):
    '''
    Overridden to allow peer certificate validation, configuration
    of SSL/ TLS version and cipher selection.  See:
    http://hg.python.org/cpython/file/c1c45755397b/Lib/httplib.py#l1144
    and ssl.wrap_socket()
    '''

    def __init__(self, host, **kwargs):
        self.ciphers = kwargs.pop('ciphers',None)
        self.ca_certs = kwargs.pop('ca_certs',None)
        self.ssl_version = kwargs.pop('ssl_version',ssl.PROTOCOL_SSLv23)

        httplib.HTTPSConnection.__init__(self,host,**kwargs)

    def connect(self):
        sock = socket.create_connection( (self.host, self.port), self.timeout )

        if self._tunnel_host:
            self.sock = sock
            self._tunnel()

        self.sock = ssl.wrap_socket( sock,
                keyfile = self.key_file,
                certfile = self.cert_file,
                ca_certs = self.ca_certs,
#                ciphers = self.ciphers,  # DOH!  This is Python 2.7-only!
                cert_reqs = ssl.CERT_REQUIRED if self.ca_certs else ssl.CERT_NONE,
                ssl_version = self.ssl_version )

client_cert_key = "%v" # file path
client_cert = "%v" #file path
ca_certs = "%v" # file path

handlers = []

if client_cert_key:
	handlers.append( HTTPSClientAuthHandler(
		key = client_cert_key,
		cert = client_cert,
		ca_certs = ca_certs,
		ssl_version = ssl.PROTOCOL_TLSv1_2,
		ciphers = 'TLS_RSA_WITH_AES_256_CBC_SHA' ) )

	http = urllib2.build_opener(*handlers)

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
	    	nomad_url = "%v"
		nomad_resp = http.open(nomad_url + "/v1/job/" + k + "-" + v) if client_cert_key else urllib2.urlopen(nomad_url + "/v1/job/" + k + "-" + v)
		status = json.loads(nomad_resp.read().decode('utf-8'))['Status']
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
		os.Getenv("NOMAD_CLIENT_KEY"), os.Getenv("NOMAD_CLIENT_CERT"),
		os.Getenv("NOMAD_CACERT"), consulHost, consulHost, nomadHost, consulHost)
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
