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
	if k := md.Get("2fa_key"); len(k) == 1 {
		key = k[0]
	}

	codes := md.Get("2fa_token")
	obj := ts2fa.NewPayload(key, op, codes...)
	return obj, nil
}

func TwoFAInterceptor(c io.ReadCloser, validator func(string, string) bool) grpc.UnaryServerInterceptor {
	// check if 2FA codes are valid.
	// todo: currently using in memory.
	var config ts2fa.Ts2FAConf

	b, rErr := ioutil.ReadAll(c)
	if rErr != nil {
		log.Println(rErr)
	}

	defer c.Close()

	if err := json.Unmarshal(b, &config); err != nil {
		log.Println(err)
	}

	config.Validator = validator
	tfa := ts2fa.New(&config)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		infoList := strings.Split(info.FullMethod, "/")

		if len(infoList) > 1 {
			obj, err := getTotpPayload(ctx, infoList[len(infoList)-1])
			if err != nil {
				log.Println(fmt.Sprintf("Error while fetching 2fa headers: %v", err))
				return nil, err
			}

			log.Printf("Validating payload %+v\n", obj)

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
