package middleware

import (
	"context"
	"fmt"
	"log"

	"github.com/mcuadros/go-version"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

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

// Check the version of the client's binary.
// Return false, if version is deprecated.
func validateVersion(cliVersion, leastVersion string) bool {
	c := version.NewConstrainGroupFromString(">=" + leastVersion)
	return c.Match(cliVersion)
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
