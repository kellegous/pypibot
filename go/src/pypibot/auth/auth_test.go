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

	if err := decodePem(key); err != nil {
		t.Fatal(err)
	}

	if err := decodePem(crt); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateCert(t *testing.T) {
	tmp, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
}
