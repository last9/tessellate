package main

import (
	"os"
	"sync"

	"google.golang.org/grpc"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const version = "0.0.1"

var (
	endpoint = kingpin.Flag("tessellate", "endpoint of Tessellate").Short('a').Default("localhost:9977").String()
)

var once sync.Once
var conn *grpc.ClientConn

func getClient() *grpc.ClientConn {
	once.Do(func() {
		con, err := grpc.Dial(*endpoint, grpc.WithInsecure())
		if err != nil {
			panic(err)
		}

		conn = con
	})

	return conn
}

func main() {
	app := kingpin.New("tessellate", "Tessellate CLI")
	app.Version(version)

	// Add your subcommand methods here.
	addWorkspaceCommand(app)
	// addLayoutCommand(app)
	addVarsCommand(app)
	addWatchCommand(app)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
