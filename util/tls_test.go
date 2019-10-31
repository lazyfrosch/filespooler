package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"testing"
	"time"
)

func createKeyPair(name string) (*os.File, *os.File) {
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal("Unable to create key: ", err)
	}

	if name == "" {
		name = "localhost"
	}

	now := time.Now()

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1337),
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             now,
		NotAfter:              now.Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{name},
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatal("Unable to create certificate: ", err)
	}

	keyFile, err := ioutil.TempFile("", "key.pem")
	if err != nil {
		log.Fatal(err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write data to key.pem: %s", err)
	}

	certFile, err := ioutil.TempFile("", "cert.pem")
	if err != nil {
		log.Fatal(err)
	}
	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: cert}); err != nil {
		log.Fatalf("Failed to write data to cert.pem: %s", err)
	}

	return keyFile, certFile
}

func cleanupFiles(files ...*os.File) {
	for _, file := range files {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}
}

func TestTlsConfig(t *testing.T) {
	cfg := &TlsConfig{}

	tlsC, err := cfg.GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	if tlsC.RootCAs != nil {
		t.Fatal("RootCAs should not be set")
	}
}

func TestTlsConfigWithCerts(t *testing.T) {
	key, cert := createKeyPair("")
	defer cleanupFiles(key, cert)

	c := cert.Name()
	k := key.Name()

	cfg := &TlsConfig{
		CAPath:   &c,
		CertPath: &c,
		KeyPath:  &k,
	}

	tlsC, err := cfg.GetConfig()
	if err != nil {
		t.Fatal(err)
	}

	if tlsC.RootCAs == nil {
		t.Fatal("RootCAs should be set")
	}
	if tlsC.Certificates == nil || len(tlsC.Certificates) != 1 {
		t.Fatal("There should be one certificate in store")
	}
}
