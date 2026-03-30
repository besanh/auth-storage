package data

import (
	"fmt"
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

// NewPrivatePEM reads the RSA private key PEM from cert/<kid>-private.pem.
func NewPrivatePEM(c *conf.Server) (biz.PrivatePEM, error) {
	filename := fmt.Sprintf("%s-private.pem", c.Kid)
	data, err := os.ReadFile(filepath.Join(certDir(), filename))
	return biz.PrivatePEM(data), err
}

// NewPublicPEM reads the RSA public key PEM from cert/<kid>-public.pem.
func NewPublicPEM(c *conf.Server) (biz.PublicPEM, error) {
	filename := fmt.Sprintf("%s-public.pem", c.Kid)
	data, err := os.ReadFile(filepath.Join(certDir(), filename))
	return biz.PublicPEM(data), err
}
