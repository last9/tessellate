package cert

import (
	"crypto/tls"

	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
)

func ServerCerts(cert, key, root string) (credentials.TransportCredentials, error) {
	// Load the Certificate KeyPair
	certificate, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot Load KeyPair")
	}

	// Create a tlsConfig
	tlsConfig := &tls.Config{
		ServerName:   "tessellate-server",
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
	}

	// If rootCert is provided, set ClientCA
	if root != "" {
		ca, err := MakeCertPool(root)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot make cert Pool")
		}

		tlsConfig.ClientCAs = ca
	}

	// Append the Credentials to the Server Options.
	return credentials.NewTLS(tlsConfig), nil
}

func ClientCerts(cert, key, root string) (credentials.TransportCredentials, error) {
	// Load the Certificate KeyPair
	certificate, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot Load KeyPair")
	}

	// Create a tlsConfig
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	// If rootCert is provided, set ClientCA
	if root != "" {
		ca, err := MakeCertPool(root)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot make cert Pool")
		}

		tlsConfig.RootCAs = ca
	}

	// Append the Credentials to the Server Options.
	return credentials.NewTLS(tlsConfig), nil
}
