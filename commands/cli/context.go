package main

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

func makeContext(ctx context.Context, totp *TwoFA) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	md := metadata.Pairs("version", version, "2fa_key", totp.Id, "2fa_token", strings.Join(totp.Codes, ","))

	ctx = metadata.NewOutgoingContext(ctx, md)

	// First Request
	// ctx := metadata.AppendToOutgoingContext(context.Background(), "version", version)
	return ctx
}
