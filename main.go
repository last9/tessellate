package main

import (
	"fmt"
	"log"
	"net"

	"github.com/tsocial/tessellate/dispatcher"
	"github.com/tsocial/tessellate/server"
	"github.com/tsocial/tessellate/storage/consul"
	"google.golang.org/grpc/reflection"
	"gopkg.in/alecthomas/kingpin.v2"
)

const Version = "0.1.2"

var (
	port      = kingpin.Flag("port", "Port no.").Short('p').Default("9977").String()
	nomadAddr = kingpin.Flag("nomad-addr", "Nomad Server Addr").
			Short('n').Default("http://127.0.0.1:4646").OverrideDefaultFromEnvar("NOMAD_ADDR").String()
	nomadHttpAuthUsername = kingpin.Flag("nomad-username", "Basic Auth Username").Envar("NOMAD_USERNAME").String()
	nomadHttpAuthPassword = kingpin.Flag("nomad-password", "Basic Auth Password").Envar("NOMAD_PASSWORD").String()
	nomadDc               = kingpin.Flag("nomad-dc", "Nomad Datacenter").Default("dc1").
				OverrideDefaultFromEnvar("NOMAD_DATACENTER").String()
	workerImage = kingpin.Flag("worker-image", "Worker Docker image name with tag").
			Default("tsl8/worker").Envar("WORKER_IMAGE").String()
	workerCPU = kingpin.Flag("worker-cpu", "Worker job CPU limit").Envar("WORKER_CPU").
			Default("200").String()
	workerMemory = kingpin.Flag("worker-memory", "Worker Memory limit in MB").Envar("WORKER_MEMORY").
			Default("200").String()
	consulAddr = kingpin.Flag("consul-addr", "Consul address").Default("127.0.0.1:8500").
			OverrideDefaultFromEnvar("CONSUL_ADDR").String()
	unlocker = "tessellate_unlock_job"
)

func main() {
	kingpin.Version(Version)
	kingpin.Parse()

	listenAddr := fmt.Sprintf("%s:%s", "0.0.0.0", *port)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := server.Grpc()
	defer s.GracefulStop()

	// Initialize Storage engine
	store := consul.MakeConsulStore(*consulAddr)
	store.Setup()

	// TODO: validate config first.
	nomadClient := dispatcher.NewNomadClient(dispatcher.NomadConfig{
		Address:    *nomadAddr,
		Username:   *nomadHttpAuthUsername,
		Password:   *nomadHttpAuthPassword,
		Datacenter: *nomadDc,
		Image:      *workerImage,
		CPU:        *workerCPU,
		Memory:     *workerMemory,
		ConsulAddr: *consulAddr,
	})

	dispatcher.Set(nomadClient)

	// check if a job exists with prefix name:
	go nomadClient.GetOrSetCleanup(unlocker)

	server.RegisterTessellateServer(s, server.New(store))

	// Register reflection service on gRPC server.
	reflection.Register(s)

	log.Printf("Serving on %v\n", listenAddr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
