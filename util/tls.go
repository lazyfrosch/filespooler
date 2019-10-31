package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

type TlsConfig struct {
	CAPath   *string
	CertPath *string
	KeyPath  *string
}

func (c *TlsConfig) GetConfig() (*tls.Config, error) {
	cfg := &tls.Config{}

	if c.CAPath != nil {
		pool := x509.NewCertPool()
		content, err := ioutil.ReadFile(*c.CAPath)
		if err != nil {
			return nil, fmt.Errorf("could not open path for reading: %s", *c.CAPath)
		}
		ok := pool.AppendCertsFromPEM(content)
		if !ok {
			return nil, fmt.Errorf("could not load certificates from: %s", *c.CAPath)
		}

		cfg.RootCAs = pool
	}

	if c.CertPath != nil && c.KeyPath != nil {
		certificate, err := tls.LoadX509KeyPair(*c.CertPath, *c.KeyPath)
		if err != nil {
			return nil, err
		}

		cfg.Certificates = []tls.Certificate{certificate}
	}

	return cfg, nil
}
