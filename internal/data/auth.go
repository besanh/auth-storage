package data

import (
	"os"
	"path/filepath"
	"runtime"

	"server/internal/biz"
	"server/internal/conf"

	"github.com/besanh/go-library/spicedb"
)

// NewSpiceClient creates a *spicedb.SpiceClient from the data config.
func NewSpiceClient(c *conf.Data) (*spicedb.SpiceClient, error) {
	opts := []spicedb.Option{}
	if c.Spicedb.Insecure {
		opts = append(opts, spicedb.WithInsecure())
	}
	return spicedb.NewClient(c.Spicedb.Endpoint, c.Spicedb.Token, opts...)
}

// certDir returns the absolute path to the cert directory next to the binary.
func certDir() string {
	_, filename, _, _ := runtime.Caller(0)
	// Walk up from internal/data → server root → cert/
	root := filepath.Join(filepath.Dir(filename), "..", "..")
	return filepath.Join(root, "cert")
}

// NewPrivatePEM reads the RSA private key PEM from cert/private_key.pem.
func NewPrivatePEM() (biz.PrivatePEM, error) {
	data, err := os.ReadFile(filepath.Join(certDir(), "private_key.pem"))
	return biz.PrivatePEM(data), err
}

// NewPublicPEM reads the RSA public key PEM from cert/public_key.pem.
func NewPublicPEM() (biz.PublicPEM, error) {
	data, err := os.ReadFile(filepath.Join(certDir(), "public_key.pem"))
	return biz.PublicPEM(data), err
}
