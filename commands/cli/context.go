package main

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

func makeContext(ctx context.Context, twoFA *TwoFA) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	md := metadata.Pairs("version", version, "2faobject", twoFA.object, "2faoperation", twoFA.operation,
		"2faident", twoFA.id, "2facodes", strings.Join(twoFA.codes, ","))

	ctx = metadata.NewOutgoingContext(ctx, md)

	// First Request
	// ctx := metadata.AppendToOutgoingContext(context.Background(), "version", version)
	return ctx
}
