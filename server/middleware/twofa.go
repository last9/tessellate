package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	ConfigFile = kingpin.Flag("2fa-config", "Config file for 2FA").Required().ExistingFile()
)

const (
	GET     = "GetLayout"
	CREATE  = "CreateLayout"
	APPLY   = "ApplyLayout"
	DEFAULT = "*"
)

type TwoFAConfig struct {
	ApplyLayout  map[string][]string `json:"ApplyLayout"`
	CreateLayout map[string][]string `json:"CreateLayout"`
}

type TwoFA struct {
	Operation string
	Id        string   `json: id`
	Codes     []string `json: codes`
}

func getTwoFA(ctx context.Context, op string) (*TwoFA, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Cannot get header metadata from context")
		return nil, errors.New("Cannot get header metadata from context")
	}
	var obj TwoFA

	if err := json.Unmarshal([]byte(md["2fa"][0]), &obj); err != nil {
		return nil, err
	}

	obj.Operation = op

	return &obj, nil
}

// todo: Update this method.
func checkCode(email, code string) (bool, error) {
	return true, nil
}

// Matches the codes per email address.
func verify2FA(obj *TwoFA, config *TwoFAConfig) error {
	var ids []string
	var exists bool
	// fetch emails for id from operation.
	switch obj.Operation {
	case APPLY:
		ids, exists = config.ApplyLayout[obj.Id]
		if !exists {
			// fetch the * email.
			ids, exists = config.ApplyLayout[DEFAULT]
			if !exists {
				// doesn't have a default 2FA enabled for this operation. Allow operation.
				return nil
			}
		}
		// At this point we expect the codes to be passed, if not passed, throw an error.
		if len(obj.Codes) == 0 {
			return errors.New("2FA codes not passed, operation not permitted.")
		}
		for i := 0; i < len(ids); i++ {
			// todo: Handle this better with passing paramters as a struct.
			ok, err := checkCode(ids[i], obj.Codes[i])
			if err != nil {
				return err
			}
			if !ok {
				return errors.New("Operation not permitted")
			}
		}
	}
	return nil
}

func TwoFAInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var config TwoFAConfig

		infoList := strings.Split(info.FullMethod, "/")
		obj, err := getTwoFA(ctx, infoList[len(infoList)-1])
		if err != nil {
			log.Println(fmt.Sprintf("Error while fetching 2fa codes: %v", err))
			return nil, err
		}

		// check if 2FA codes are valid.
		// todo: currently using in memory.
		b, rErr := ioutil.ReadFile(*ConfigFile)
		if rErr != nil {
			return nil, rErr
		}

		if err := json.Unmarshal(b, &config); err != nil {
			return nil, err
		}

		// todo: Some restructuring of code logic needed here. Handled using switch cases for now.
		if err := verify2FA(obj, &config); err != nil {
			return nil, err
		}

		// this operation never expects a 2FA for the object, allow the operation to be performed.
		resp, err := handler(ctx, req)
		if err != nil {
			fmt.Printf("%+v", err)
			return nil, err
		}

		// Return handler's response and err.
		return resp, nil
	}
}
