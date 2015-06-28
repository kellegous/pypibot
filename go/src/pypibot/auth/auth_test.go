package auth

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
)

const keyBits = 2048

func assertValidCrtAndKey(t *testing.T, crtPem, keyPem *pem.Block) {
	if crtPem.Type != "CERTIFICATE" {
		t.Fatalf("Expected pem type of CERTIFICATE, got %s", crtPem.Type)
	}

	_, err := x509.ParseCertificate(crtPem.Bytes)
	if err != nil {
		t.Fatal(err)
	}

	if keyPem.Type != "RSA PRIVATE KEY" {
		t.Fatalf("Expected pem type of RSA PRIVATE KEY, got %s", keyPem.Type)
	}

	_, err = x509.ParsePKCS1PrivateKey(keyPem.Bytes)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateServerCert(t *testing.T) {
	crtPem, keyPem, err := GenerateServerCert(keyBits, "kellegous")
	if err != nil {
		t.Fatal(err)
	}

	assertValidCrtAndKey(t, crtPem, keyPem)
}

func TestGenerateClientCert(t *testing.T) {
	caCrtPem, caKeyPem, err := GenerateServerCert(keyBits, "kellegous")
	if err != nil {
		t.Fatal(err)
	}

	crtPem, keyPem, err := GenerateClientCert(keyBits, caCrtPem, caKeyPem)
	if err != nil {
		t.Fatal(err)
	}

	// TODO(knorton): Assert that client is signed by server.
	assertValidCrtAndKey(t, crtPem, keyPem)
}
