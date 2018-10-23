package main

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func makeContext(ctx context.Context, totp *TwoFA) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	md := metadata.Pairs("version", version)

	if totp != nil {
		md.Set("2fa_key", totp.Id)
		md.Set("2fa_token", totp.Codes...)
	}

	ctx = metadata.NewOutgoingContext(ctx, md)

	// First Request
	// ctx := metadata.AppendToOutgoingContext(context.Background(), "version", version)
	return ctx
}
