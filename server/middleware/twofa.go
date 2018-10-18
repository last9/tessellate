package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile = kingpin.Flag("2fa-config", "Config file for 2FA").
		Default("/home/talina06/workspace/gopath/src/github.com/tsocial/tessellate/server/2fa.json").ExistingFile()
)

const (
	LAYOUT    = "layout"
	WORKSPACE = "workspace"

	GET    = "get"
	CREATE = "create"
	APPLY  = "apply"
)

type TwoFAConfig struct {
	Layout    []string            `json:"layout"`
	Workspace []string            `json:"workspace"`
	Apply     map[string][]string `json:"apply"`
	Create    map[string][]string `json:"create"`
}

type TwoFA struct {
	Object    string   `json: object`
	Operation string   `json: operation`
	Id        string   `json: id`
	Codes     []string `json: codes`
}

func getTwoFA(ctx context.Context) (*TwoFA, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Cannot get header metadata from context")
		return nil, errors.New("Cannot get header metadata from context")
	}

	if md["2fa"] == nil || len(md["2fa"]) == 0 {
		return nil, errors.New("2FA Object not found in the header")
	}

	var obj TwoFA

	if err := json.Unmarshal([]byte(md["2fa"][0]), &obj); err != nil {
		return nil, err
	}

	return &obj, nil
}

func contains(items []string, match string) bool {
	for _, v := range items {
		if v == match {
			return true
		}
	}
	return false
}

func checkCode(email, code string) (bool, error) {
	return true, nil
}

func TwoFAInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var config TwoFAConfig
		obj, err := getTwoFA(ctx)
		if err != nil {
			log.Println(fmt.Sprintf("Error while fetching 2fa codes: %v", err))
			return nil, err
		}

		// check if 2FA codes are valid.
		// todo: currently using in memory, tomorrow use a more efficient store.
		b, rErr := ioutil.ReadFile(*configFile)
		if rErr != nil {
			return nil, rErr
		}

		if err := json.Unmarshal(b, &config); err != nil {
			return nil, err
		}

		// check if 2fa exists for object and operation.
		switch obj.Object {
		case LAYOUT:
			if contains(config.Layout, obj.Operation) {
				fmt.Println("2FA is enabled.")
			} else {
				return nil, nil
			}
		}

		// fetch emails for id from operation.
		switch obj.Operation {
		case APPLY:
			ids := config.Apply[obj.Id]
			for i := 0; i < len(ids); i++ {

				// todo: Handle this better with passing paramters as a struct.
				ok, err := checkCode(ids[i], obj.Codes[i])
				if err != nil {
					return nil, err
				}
				if !ok {
					return nil, err
				}
			}
		}

		// call
		return nil, nil
	}
}
