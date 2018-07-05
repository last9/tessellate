package main

import (
	"fmt"
	"log"
	"net"

	server "github.com/oogway/tessellate/server"
	"github.com/oogway/tessellate/storage/consul"
	"google.golang.org/grpc/reflection"
	"gopkg.in/alecthomas/kingpin.v2"
)

const Version = "0.0.1"

var port = kingpin.Flag("port", "Port no.").Short('p').Default("9977").String()

func main() {
	kingpin.Version(Version)
	kingpin.Parse()

	listenAddr := fmt.Sprintf("%s:%s", "0.0.0.0", *port)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpcServer()
	defer s.GracefulStop()

	store := consul.MakeConsulStore()
	store.Setup()

	server.RegisterTessellateServer(s, server.New(store))

	// Register reflection service on gRPC server.
	reflection.Register(s)

	log.Printf("Serving on %v\n", listenAddr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
