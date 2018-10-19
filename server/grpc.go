package server

import (
	"io"
	"log"

	"github.com/pquerna/otp/totp"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/tsocial/tessellate/cert"
	"github.com/tsocial/tessellate/fault"
	"github.com/tsocial/tessellate/server/middleware"
	"google.golang.org/grpc"
	"gopkg.in/alecthomas/kingpin.v2"
)

const DefaultVersion = "0.1.0"

var (
	rootCert = kingpin.Flag("root-cert-file", "Root Cert File").Envar("ROOT_CERT_FILE").
			String()
	certFile = kingpin.Flag("cert-file", "Cert File").Envar("CERT_FILE").String()
	keyFile  = kingpin.Flag("key-file", "Key File").Envar("KEY_FILE").String()
	support  = (kingpin.Flag("least-cli-version", "Client's least supported version by Tessellate.")).
			Default(DefaultVersion).OverrideDefaultFromEnvar("LEAST_CLI_VERSION").String()
	twoFAConfig = io.ReadCloser(*kingpin.Flag("totp-config", "Config file for 2FA").Required().File())
)

func customFunc(t interface{}) error {
	return fault.Printer(t)
}

var validator = totp.Validate

func Grpc() *grpc.Server {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(customFunc),
	}

	unaries := []grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(opts...),
		middleware.UnaryServerInterceptor(*support),
		middleware.TwoFAInterceptor(twoFAConfig, validator),
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
