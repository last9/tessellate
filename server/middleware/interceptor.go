package middleware

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"strings"
)

func getVersionId(ctx context.Context) (string, error) {
	headers, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("Cannot get header metadata from context")
		panic(errors.New("Cannot get header metadata from context"))
	}

	log.Println("headers are")
	log.Println(headers)
	if headers["version"] != nil {
		version := headers["version"][0]
		return version, nil
	}
	return "", errors.New("Version not found in the header.")
}

// Check the version of the client's binary.
// Return false, if version is deprecated.
func validateVersion(version, latestVersion string) bool {
	return strings.EqualFold(version, latestVersion)
}

func UnaryServerInterceptor(supportVersion string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Get the version from the header.
		version, err := getVersionId(ctx)

		// If the id is empty, return a older version error.
		if version == "" {
			log.Printf("Version not found.")
			return nil, errors.New("You are using an older version of Tessellate. Download the new version from here.")
		}
		if !validateVersion(version, supportVersion) {
			return nil, errors.New("You are using an older version of Tessellate. Download the new version from here.")
		}

		// Else, pass the request ahead to the handler.
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		// Return handler's response and err.
		return resp, err
	}
}
