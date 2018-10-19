package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func TestInterceptor(t *testing.T) {
	port := 59999
	go func() {
		*support = "0.0.4"

		listenAddr := fmt.Sprintf(":%v", port)
		lis, err := net.Listen("tcp", listenAddr)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := Grpc()
		defer s.GracefulStop()

		RegisterTessellateServer(s, New(store))

		// Register reflection service on gRPC server.
		reflection.Register(s)

		log.Printf("Serving on %v\n", listenAddr)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%v", port), opts...)
	if err != nil {
		panic(err)
	}

	tClient := NewTessellateClient(conn)

	t.Run("Should raise an error for non supported lower versions.", func(t *testing.T) {
		// Get a client instance with new version passed in metadata.
		// First Request
		ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.1")
		log.Printf("Context: %+v", ctx)

		resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: "test"})

		// Assert for failure.
		assert.NotEmpty(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Valid version. Should forward request to server and return a successful response.", func(t *testing.T) {
		// 1. Pass same versions.
		ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.6")
		// log.Printf("Context: %+v", ctx)

		resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: "test"})

		// 2. Assert for success.
		assert.Nil(t, err)
		assert.Equal(t, resp, &Ok{})
	})

	t.Run("Boundary case: Should pass for the exact version support.", func(t *testing.T) {
		// 1. Pass same versions.
		ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.4")
		// log.Printf("Context: %+v", ctx)

		resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: "test"})

		// 2. Assert for success.
		assert.Nil(t, err)
		assert.Equal(t, resp, &Ok{})
	})
}
