package main

import (
	"log"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/oogway/tessellate/fault"
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
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(unaries...)),
	)

	return s
}
