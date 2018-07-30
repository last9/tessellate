package main

import (
	"log"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/tsocial/tessellate/cert"
	"github.com/tsocial/tessellate/fault"
	"github.com/tsocial/tessellate/server/middleware"
	"google.golang.org/grpc"
)

func customFunc(t interface{}) error {
	return fault.Printer(t)
}

func grpcServer() *grpc.Server {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(customFunc),
	}

	unaries := []grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(opts...),
		middleware.UnaryServerInterceptor(*support),
	}

	sopts := []grpc.ServerOption{}

	if *certFile != "" && *keyFile != "" {
		creds, err := cert.ServerCerts(*certFile, *keyFile, *rootCert)
		if err != nil {
			panic(err)
		}

		// Append the Credentials to the Server Options.
		sopts = append(sopts, grpc.Creds(creds))
	}

	sopts = append(sopts, grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaries...)))

	return grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaries...)),
	)
}
