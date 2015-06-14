package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// Use openssl to generate a new RSA key.
func GenerateRsaKey(filename string, size int) error {
	if err := exec.Command(
		"openssl",
		"genrsa",
		"-out",
		filename,
		fmt.Sprintf("%d", size)).Run(); err != nil {
		return fmt.Errorf("unable to create rsa key: %s", err)
	}

	return nil
}

// Use openssl to generate a new private CA certificate and key.
func GenerateCa(name, crtFile, keyFile string) error {
	if err := GenerateRsaKey(keyFile, 2048); err != nil {
		return err
	}

	if err := exec.Command(
		"openssl",
		"req",
		"-x509",
		"-new",
		"-key", keyFile,
		"-out", crtFile,
		"-days", "730",
		"-subj", fmt.Sprintf("/CN=\"%s\"", name)).Run(); err != nil {
		return fmt.Errorf("unable to create cert: %s", err)
	}

	return nil
}

func closeAndRemove(f *os.File) {
	f.Close()
	os.Remove(f.Name())
}

// Use openssl to generate a new certificate signed by the given CA key.
func GenerateCert(host, caCrtFile, caKeyFile, crtFile, keyFile string) (string, error) {
	if err := GenerateRsaKey(keyFile, 2048); err != nil {
		return "", err
	}

	tmp, err := ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)

	reqFile := filepath.Join(tmp, "req")
	serFile := filepath.Join(tmp, "ser")

	if err := exec.Command(
		"openssl",
		"req",
		"-new",
		"-out", reqFile,
		"-key", keyFile,
		"-subj", fmt.Sprintf("/CN=%s", host)).Run(); err != nil {
		return "", fmt.Errorf("unable to create signing request: %s", err)
	}

	if err := exec.Command(
		"openssl",
		"x509",
		"-req",
		"-in", reqFile,
		"-out", crtFile,
		"-CAkey", caKeyFile,
		"-CA", caCrtFile,
		"-days", "365",
		"-CAcreateserial",
		"-CAserial", serFile).Run(); err != nil {
		return "", fmt.Errorf("unable to create certificate: %s", err)
	}

	b, err := ioutil.ReadFile(serFile)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func parsePrivateKey(b []byte) (*rsa.PrivateKey, error) {
	x, err := x509.ParsePKCS1PrivateKey(b)
	if err == nil {
		return x, nil
	}

	y, err := x509.ParsePKCS8PrivateKey(b)
	if err == nil {
		return y.(*rsa.PrivateKey), nil
	}

	return nil, errors.New("unable to decode private key")
}

func ReadPrivateKey(filename string) (*rsa.PrivateKey, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	p, _ := pem.Decode(b)
	if p == nil {
		return nil, fmt.Errorf("invalid pem: %s", filename)
	}

	return parsePrivateKey(p.Bytes)
}

func GetPublicKey(prv *rsa.PrivateKey) ([]byte, error) {
	return x509.MarshalPKIXPublicKey(&prv.PublicKey)
}
