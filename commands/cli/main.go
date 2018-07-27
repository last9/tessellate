package main

import (
	"log"
	"os"
	"sync"

	"gitlab.com/tsocial/sre/tessellate/cert"
	"gitlab.com/tsocial/sre/tessellate/server"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "0.0.1"

var (
	endpoint *string
	certFile *string
	keyFile  *string
	rootCert *string
)

var once sync.Once

func getClient() server.TessellateClient {
	var client server.TessellateClient

	once.Do(func() {

		opts := []grpc.DialOption{}
		if *certFile != "" && *keyFile != "" {
			creds, err := cert.ClientCerts(*certFile, *keyFile, *rootCert)
			if err != nil {
				panic(err)
			}
			opts = append(opts, grpc.WithTransportCredentials(creds))
		} else {
			opts = append(opts, grpc.WithInsecure())
		}

		conn, err := grpc.Dial(*endpoint, opts...)
		if err != nil {
			panic(err)
		}

		client = server.NewTessellateClient(conn)
	})

	return client
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	app := kingpin.New("tessellate", "Tessellate CLI")

	rootCert = app.Flag("root-cert", "Root Cert File").String()
	certFile = app.Flag("cert-file", "Cert File").String()
	keyFile = app.Flag("key-file", "Key File").String()

	endpoint = app.Flag("tessellate", "endpoint of Tessellate").Short('a').Default("localhost:9977").String()
	app.Version(version)

	// Add your command methods here.
	addWorkspaceCommand(app)
	addLayoutCommands(app)
	addWatchCommand(app)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
