package main

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	gw "github.com/tsocial/tessellate/server"
)

const Version = "0.0.1"

var (
	endpoint = kingpin.Flag("service_addr", "endpoint of YourService").Short('a').Default("localhost:9977").String()
	port     = kingpin.Flag("port", "Port no.").Short('p').Default("8080").String()
)

func run() error {
	kingpin.Version(Version)
	kingpin.Parse()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw.RegisterTessellateHandlerFromEndpoint(ctx, mux, *endpoint, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(fmt.Sprintf(":%v", *port), mux)
}

func main() {
	defer glog.Flush()

	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
