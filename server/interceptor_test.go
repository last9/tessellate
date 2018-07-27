package server

/*import (
	"testing"
	"os"
	"gitlab.com/tsocial/sre/tessellate/storage/consul"
	"gitlab.com/tsocial/sre/tessellate/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"context"
)

var cStore storage.Storer
var tServer TessellateServer
var tClient TessellateClient

func TestInterceptor_GetAndCheckVersion(t *testing.T) {

	// 1. Get a server instance with old version

	store = consul.MakeConsulStore(os.Getenv("CONSUL"))
	store.Setup()

	server = New(store)

	// 2. Get a client instance with new version passed in metadata.
	// 3. assert for failure.

	// 1. Pass same versions.
	// 2. Assert for success.


	t.Run("Should raise an error for version mismatch", func(t *testing.T) {

		opts := []grpc.DialOption{}

		conn, err := grpc.Dial("127.0.0.1:9977", opts...)
		if err != nil {
			panic(err)
		}

		tClient = NewTessellateClient(conn)

		// First Request
		ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "1")
		log.Printf("Context: %+v", ctx)


	})

	t.Run("Valid version. Should forward request to server and return a successful response.", func(t *testing.T) {

		opts := []grpc.DialOption{}

		conn, err := grpc.Dial("127.0.0.1:9977", opts...)
		if err != nil {
			panic(err)
		}

		tClient = NewTessellateClient(conn)

		// First Request
		ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "1")
		log.Printf("Context: %+v", ctx)
	})
}*/
