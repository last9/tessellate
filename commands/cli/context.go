package main

import (
	"context"
	"google.golang.org/grpc/metadata"
	"log"
)

func makeContext(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	md := metadata.Pairs("version", version)
	ctx = metadata.NewOutgoingContext(ctx, md)

	// First Request
	// ctx := metadata.AppendToOutgoingContext(context.Background(), "version", version)
	log.Printf("Context: %+v", ctx)
	return ctx
}
