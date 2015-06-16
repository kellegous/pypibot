package store

import (
	"crypto/tls"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func assertValidCert(t *testing.T, crt, key string) {
	if _, err := tls.LoadX509KeyPair(crt, key); err != nil {
		t.Fatal(err)
	}
}

func TestCreate(t *testing.T) {
	tmp, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	path := filepath.Join(tmp, "data")

	if err := Create(path); err != nil {
		t.Fatal(err)
	}

	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	caCrt, caKey := s.CaCertFiles()
	assertValidCert(t, caCrt, caKey)

	rpcCrt, rpcKey := s.RpcCertFiles()
	assertValidCert(t, rpcCrt, rpcKey)
}
