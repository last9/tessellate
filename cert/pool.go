package cert

import (
	"crypto/x509"
	"io/ioutil"

	"github.com/pkg/errors"
)

// MakeCertPool generates a CertPool based on rootPath.
func MakeCertPool(rootPath string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	bs, err := ioutil.ReadFile(rootPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read client ca cert.")
	}

	ok := certPool.AppendCertsFromPEM(bs)
	if !ok {
		return nil, errors.New("Failed to append client certs")
	}

	return certPool, nil
}
