package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/tsocial/ts2fa/otp"

	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func getTotpPayload(ctx context.Context, op string) (*ts2fa.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Cannot get header metadata from context")
		return nil, errors.New("Cannot get header metadata from context")
	}

	var key string
	if k, ok := md["2fa_key"]; ok && len(k) == 1 {
		key = k[0]
	}

	codes := []string{}
	if c, ok := md["2fa_token"]; ok && len(c) == 1 {
		codes = strings.Split(c[0], ",")
	}

	obj := ts2fa.NewPayload(op, key, codes...)
	return obj, nil
}

func TwoFAInterceptor(c io.ReadCloser, validator func(string, string) bool) grpc.UnaryServerInterceptor {
	// check if 2FA codes are valid.
	// todo: currently using in memory.
	var rules ts2fa.Rules

	b, rErr := ioutil.ReadAll(c)
	if rErr != nil {
		log.Println(rErr)
	}

	defer c.Close()

	if err := json.Unmarshal(b, &rules); err != nil {
		log.Println(err)
	}

	tfa := ts2fa.New(&ts2fa.Ts2FAConf{
		Rules:     rules,
		Validator: validator,
	})

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		infoList := strings.Split(info.FullMethod, "/")

		if len(infoList) > 1 {
			obj, err := getTotpPayload(ctx, infoList[len(infoList)-1])
			if err != nil {
				log.Println(fmt.Sprintf("Error while fetching 2fa headers: %v", err))
				return nil, err
			}

			valid, err := tfa.Verify(obj)
			if err != nil {
				return nil, err
			}

			if !valid {
				return nil, fmt.Errorf("totp Validation failed")
			}
		}

		// this operation never expects a 2FA for the object, allow the operation to be performed.
		return handler(ctx, req)
	}
}
