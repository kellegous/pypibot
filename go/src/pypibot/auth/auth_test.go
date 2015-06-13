package auth

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func decodePem(filename string) error {
	d, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	b, _ := pem.Decode(d)
	if b == nil {
		return fmt.Errorf("pem decode failed on %s", filename)
	}

	return nil
}

func decodeAll(t *testing.T, filenames ...string) {
	for _, filename := range filenames {
		if err := decodePem(filename); err != nil {
			t.Fatal(err)
		}
	}
}

func TestGenerateCa(t *testing.T) {
	tmp, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmp)

	key := filepath.Join(tmp, "key.pem")
	crt := filepath.Join(tmp, "crt.pem")

	if err := GenerateCa("kellegous ltd", crt, key); err != nil {
		t.Fatal(err)
	}

	decodeAll(t, key, crt)
}

func TestGenerateCert(t *testing.T) {
	tmp, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	caKey := filepath.Join(tmp, "ca-key.pem")
	caCrt := filepath.Join(tmp, "ca-crt.pem")

	if err := GenerateCa("kellegous ltd", caCrt, caKey); err != nil {
		t.Fatal(err)
	}

	crt := filepath.Join(tmp, "crt.pem")
	key := filepath.Join(tmp, "key.pem")

	ser, err := GenerateCert("kellego.us", caCrt, caKey, crt, key)
	if err != nil {
		t.Fatal(err)
	}

	if ser == "" {
		t.Fatal("empty serial")
	}

	decodeAll(t, caCrt, caKey, crt, key)
}
