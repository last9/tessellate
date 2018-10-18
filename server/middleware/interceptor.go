package middleware

import (
	"context"
	"fmt"
	"log"

	"strings"

	"io/ioutil"

	"encoding/json"

	"github.com/mcuadros/go-version"
	"github.com/pkg/errors"
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

type twoFAConfig struct {
	Layout    []string            `json:"layout"`
	Workspace []string            `json:"workspace"`
	Apply     map[string][]string `json:"apply"`
	Create    map[string][]string `json:"create"`
}

type twoFA struct {
	object    string
	operation string
	id        string
	codes     []string
}

func NewTwoFA(object, operation, id string, codes []string) *twoFA {
	return &twoFA{
		object:    object,
		operation: operation,
		id:        id,
		codes:     codes,
	}
}

func getVersionId(ctx context.Context) (string, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Cannot get header metadata from context")
		return "", errors.New("Cannot get header metadata from context")
	}

	if headers["version"] == nil || len(headers["version"]) == 0 {
		return "", errors.New("Version not found in the header")
	}
	return headers["version"][0], nil
}

func getTwoFA(ctx context.Context) (*twoFA, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Cannot get header metadata from context")
		return nil, errors.New("Cannot get header metadata from context")
	}

	if md["2faobject"] == nil || len(md["2faobject"]) == 0 {
		return nil, errors.New("2FA Object not found in the header")
	}
	if md["2faoperation"] == nil || len(md["2faoperation"]) == 0 {
		return nil, errors.New("2FA Operation not found in the header")
	}
	if md["2faident"] == nil || len(md["2faident"]) == 0 {
		return nil, errors.New("2FA Ident not found in the header")
	}
	if md["2facodes"] == nil || len(md["2facodes"]) < 1 {
		return nil, errors.New("2FA Ident not found in the header")
	}

	return NewTwoFA(md["2faobject"][0], md["2faoperation"][0], md["2faident"][0], strings.Split(md["2facodes"][0], ",")), nil
}

// Check the version of the client's binary.
// Return false, if version is deprecated.
func validateVersion(cliVersion, leastVersion string) bool {
	c := version.NewConstrainGroupFromString(">=" + leastVersion)
	return c.Match(cliVersion)
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
		var config twoFAConfig
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
		switch obj.object {
		case LAYOUT:
			if contains(config.Layout, obj.operation) {
				fmt.Println("2FA is enabled.")
			} else {
				return nil, nil
			}
		}

		fmt.Println(obj.codes)
		// fetch emails for id from operation.
		switch obj.operation {
		case APPLY:
			ids := config.Apply[obj.id]
			for i := 0; i < len(ids); i++ {

				// todo: Handle this better with passing paramters as a struct.
				_, err = checkCode(ids[i], obj.codes[i])
				if err != nil {
					return nil, err
				}
			}
		}

		// call
		return nil, nil
	}
}

func UnaryServerInterceptor(supportVersion string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		url := "https://github.com/tsocial/tessellate/releases"

		// Get the version from the header.
		version, err := getVersionId(ctx)
		if err != nil {
			return nil, err
		}

		versionErr := errors.Errorf(
			"You are using an older version: %v of Tessellate CLI. "+
				"Download the newer version (>= %v) from: %v",
			version, supportVersion, url)

		// If the id is empty, return a older version error.
		if version == "" {
			log.Printf("Version not found.")
			return nil, versionErr
		}

		if !validateVersion(version, supportVersion) {
			return nil, versionErr
		}

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
