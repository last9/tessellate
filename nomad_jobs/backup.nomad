# consul-snapshot nomad jobfile
job "consul-snapshot" {
  datacenters = [ "dc1" ]
  type = "batch"

  periodic {
    cron             = "*/1 * * * * *"
    prohibit_overlap = true
  }

  group "consul-snapshot" {
    count = 1

    task "backup" {
      driver = "raw_exec"
      config {
        command = "bash"
        args = ["-exc", "ts=`date +%s`; consul snapshot save -http-addr=http://control.ha.tsengineering.io:8500 $ts+=_backup.snap"]
      }

      resources {
        cpu    = 100
        memory = 256
      }

    }
  }
}
