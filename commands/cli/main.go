package main

import (
	"log"
	"os"
	"sync"

	"github.com/tsocial/tessellate/cert"
	"github.com/tsocial/tessellate/server"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "0.2.1"

var (
	endpoint *string
	certFile *string
	keyFile  *string
	rootCert *string
	codes    *[]string
)

var once sync.Once
var client server.TessellateClient

type TwoFA struct {
	Id    string
	Codes []string
}

func NewTwoFA(id string, codes []string) *TwoFA {
	return &TwoFA{
		Id:    id,
		Codes: codes,
	}
}

func getClient() server.TessellateClient {
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
	app := kingpin.New("tsl8", "Tessellate CLI")

	rootCert = app.Flag("root-cert", "Root Cert File").String()
	certFile = app.Flag("cert-file", "Cert File").String()
	keyFile = app.Flag("key-file", "Key File").String()
	codes = app.Flag("2fa", "2FA code of the stakeholder").Strings()

	endpoint = app.Flag("tsl8_server", "endpoint of Tessellate Server").Short('a').Envar("TESSELLATE_SERVER").
		Default("localhost:9977").String()

	app.Version(version)

	// Add your command methods here.
	addWorkspaceCommand(app)
	addLayoutCommands(app)
	addWatchCommand(app)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
