package main

import (
	"log"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"gitlab.com/tsocial/sre/tessellate/cert"
	"gitlab.com/tsocial/sre/tessellate/fault"
	"gitlab.com/tsocial/sre/tessellate/server/middleware"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
)

func customFunc(t interface{}) error {
	return fault.Printer(t)
}

var (
	rootCert = kingpin.Flag("root-cert", "Root Cert File").String()
	certFile = kingpin.Flag("cert-file", "Cert File").String()
	keyFile  = kingpin.Flag("key-file", "Key File").String()
	support  = kingpin.Flag("support-version", "Client's least supported version by Tessellate.").String()
)

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
