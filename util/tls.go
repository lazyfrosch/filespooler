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
		if !pool.AppendCertsFromPEM(content) {
			return nil, fmt.Errorf("could not load certificates from: %s", *c.CAPath)
		}

		cfg.RootCAs = pool
		clientPool := *pool
		cfg.ClientCAs = &clientPool
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

func GetNamesFromCertificate(cert *x509.Certificate) []string {
	var names []string
	if cert.Subject.CommonName != "" {
		names = append(names, cert.Subject.CommonName)
	}

	for _, name := range cert.DNSNames {
		names = append(names, name)
	}

	return names
}

func ValidateNamesOnCertificate(cert *x509.Certificate, whitelist []string) (bool, string) {
	certNames := GetNamesFromCertificate(cert)

	for _, wlName := range whitelist {
		for _, certName := range certNames {
			if wlName == certName {
				return true, wlName
			}
		}
	}

	return false, ""
}
