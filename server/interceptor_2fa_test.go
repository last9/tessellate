package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"testing"

	"github.com/tsocial/ts2fa/otp"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

func Test2FAInterceptor(t *testing.T) {
	port := 58999

	pre_secret, pre_token, v := ts2fa.TestValidator(validator)
	validator = v

	x := ts2fa.Ts2FAConf{
		Rules: ts2fa.Rules{
			"blink_staging": {
				"SaveWorkspace": []string{pre_secret},
				"*":             []string{},
			},
			"dj_staging": {
				"*": []string{},
			},
			"*": {
				"*": []string{"some-secret"},
			},
		},
	}

	b, err := json.Marshal(x)
	assert.Nil(t, err)

	twofaIO = ioutil.NopCloser(bytes.NewBuffer(b))

	go func() {
		*support = "0.0.1"

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
		err = s.Serve(lis)
		assert.Nil(t, err, fmt.Sprintf("failed to serve: %v", err))
	}()

	opts := []grpc.DialOption{}
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%v", port), opts...)
	assert.Nil(t, err)

	tClient := NewTessellateClient(conn)

	t.Run("Missing token in Context", func(t *testing.T) {
		ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.1")
		resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: "test"})

		assert.Contains(t, err.Error(), "invalid no. of Tokens passed. Expected 1 got 0")
		// Assert for failure.
		assert.NotEmpty(t, err)
		assert.Nil(t, resp)
	})

	t.Run("Key that doesn't need a token", func(t *testing.T) {
		t.Run("Missing context", func(t *testing.T) {
			wid := "dj_staging"
			ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.1")
			resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: wid})
			assert.Contains(t, err.Error(), "invalid no. of Tokens passed. Expected 1 got 0")
			assert.NotEmpty(t, err)
			assert.Nil(t, resp)
		})

		t.Run("Valid context", func(t *testing.T) {
			wid := "dj_staging"
			ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.1", "2fa_key", wid)
			resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: wid})

			assert.Nil(t, err)
			assert.Equal(t, resp, &Ok{})
		})

	})

	t.Run("Incorrect token", func(t *testing.T) {
		t.Run("Specific Id", func(t *testing.T) {
			wid := "blink_staging"
			ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.1", "2fa_key", wid, "2fa_token", "123456")
			resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: wid})

			assert.Contains(t, err.Error(), "validation failed for OTP")
			// Assert for failure.
			assert.NotEmpty(t, err)
			assert.Nil(t, resp)
		})

		t.Run("Catch-all Rule", func(t *testing.T) {
			wid := "test_staging"
			ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.1", "2fa_key", wid, "2fa_token", "123456")
			resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: wid})

			assert.Contains(t, err.Error(), "validation failed for OTP")
			// Assert for failure.
			assert.NotEmpty(t, err)
			assert.Nil(t, resp)
		})
	})

	t.Run("Should Pass for valid tokens", func(t *testing.T) {
		wid := "blink_staging"
		ctx := metadata.AppendToOutgoingContext(context.Background(), "version", "0.0.1", "2fa_key", wid, "2fa_token", pre_token)
		resp, err := tClient.SaveWorkspace(ctx, &SaveWorkspaceRequest{Id: wid})

		t.Log(resp, err)
	})
}
