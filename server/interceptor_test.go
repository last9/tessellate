package server

import (
	"context"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"testing"
)

var tClient TessellateClient

func TestInterceptor_GetAndCheckVersion(t *testing.T) {
	t.Run("Should raise an error for non supported lower versions.", func(t *testing.T) {
		// Get a client instance with new version passed in metadata.
		opts := []grpc.DialOption{}

		opts = append(opts, grpc.WithInsecure())

		conn, err := grpc.Dial("127.0.0.1:9977", opts...)
		if err != nil {
			panic(err)
		}

		tClient = NewTessellateClient(conn)

		// First Request
		ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.1")
		log.Printf("Context: %+v", ctx)

		resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: "test"})

		// Assert for failure.
		assert.NotEmpty(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Valid version. Should forward request to server and return a successful response.", func(t *testing.T) {

		opts := []grpc.DialOption{}

		opts = append(opts, grpc.WithInsecure())

		conn, err := grpc.Dial("127.0.0.1:9977", opts...)
		if err != nil {
			panic(err)
		}

		tClient = NewTessellateClient(conn)

		// 1. Pass same versions.
		ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.6")
		// log.Printf("Context: %+v", ctx)

		resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: "test"})

		// 2. Assert for success.
		assert.Nil(t, err)
		assert.Equal(t, resp, &Ok{})
	})
}
