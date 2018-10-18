package main

import (
	"context"
	"encoding/json"
	"log"

	"google.golang.org/grpc/metadata"
)

func makeContext(ctx context.Context, totp *twoFA) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	b, err := json.Marshal(totp)
	if err != nil {
		log.Println("Error while marshalling twoFA struct to json.")
		return nil
	}

	md := metadata.Pairs("version", version, "2fa", string(b))

	ctx = metadata.NewOutgoingContext(ctx, md)

	// First Request
	// ctx := metadata.AppendToOutgoingContext(context.Background(), "version", version)
	return ctx
}
