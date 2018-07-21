package main

import (
	"context"
	"log"

	"github.com/tsocial/tessellate/server"
	"google.golang.org/grpc"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const version = "0.0.1"

var (
	endpoint = kingpin.Flag("service_addr", "endpoint of YourService").Short('a').Default("localhost:9977").String()
)

func main() {
	kingpin.Version(version)
	kingpin.Parse()

	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	req := server.GetWorkspaceRequest{
		Id: "hello",
	}

	client := server.NewTessellateClient(conn)
	w, err := client.GetWorkspace(context.Background(), &req)
	log.Println(w)
	log.Println(err)
}
