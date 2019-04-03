package middleware

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/getsentry/raven-go"
	"google.golang.org/grpc"
)

func SentryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		defer func() {
			if err := recover(); err != nil {
				debug.PrintStack()
				raven.CaptureErrorAndWait(errors.New(fmt.Sprint(err)), nil)
			}
		}()

		// Else, pass the request ahead to the handler.
		resp, err := handler(ctx, req)
		if err != nil {
			fmt.Printf("%+v", err)
			return nil, err
		}

		// Return handler's response and err.
		return resp, nil
	}
}
