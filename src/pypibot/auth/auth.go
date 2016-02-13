package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"
)

const (
	certInfoCountry = "United States"
	certInfoOrgName = "kellegous"
	certInfoOrgUnit = "pypibot"
	expiresAfter    = 1000 * 24 * time.Hour
)

func GenerateServerCert(bits int, host string) (*pem.Block, *pem.Block, error) {
	prv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	tpl := &x509.Certificate{
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(expiresAfter),
		SerialNumber:          sn,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		KeyUsage:              x509.KeyUsageDataEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		DNSNames:              []string{host},
		Subject: pkix.Name{
			Country:            []string{certInfoCountry},
			Organization:       []string{certInfoOrgName},
			OrganizationalUnit: []string{certInfoOrgUnit},
		},
	}

	crt, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &prv.PublicKey, prv)
	if err != nil {
		return nil, nil, err
	}

	return toPems(crt, prv)
}

func GenerateClientCert(bits int, caCrtPem, caKeyPem *pem.Block) (*pem.Block, *pem.Block, error) {
	prv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	sn, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	caCrt, err := x509.ParseCertificate(caCrtPem.Bytes)
	if err != nil {
		return nil, nil, err
	}

	caKey, err := x509.ParsePKCS1PrivateKey(caKeyPem.Bytes)
	if err != nil {
		return nil, nil, err
	}

	tpl := &x509.Certificate{
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(expiresAfter),
		SerialNumber:          sn,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		Subject: pkix.Name{
			Country:            []string{certInfoCountry},
			Organization:       []string{certInfoOrgName},
			OrganizationalUnit: []string{certInfoOrgUnit},
		},
	}

	crt, err := x509.CreateCertificate(rand.Reader, tpl, caCrt, &prv.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	return toPems(crt, prv)
}

func toPems(crt []byte, key *rsa.PrivateKey) (*pem.Block, *pem.Block, error) {
	return &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: crt,
		}, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		}, nil
}

func writePem(b *pem.Block, filename string) error {
	w, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer w.Close()

	return pem.Encode(w, b)
}

func WriteBothPems(crt *pem.Block, crtFile string, key *pem.Block, keyFile string) error {
	if err := writePem(crt, crtFile); err != nil {
		return err
	}

	if err := writePem(key, keyFile); err != nil {
		return err
	}

	return nil
}

func ReadPem(filename string) (*pem.Block, error) {
	c, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	b, _ := pem.Decode(c)
	if b == nil {
		return nil, fmt.Errorf("unable to decode pem in %s", filename)
	}

	return b, nil
}

func ReadBothPems(crtFile, keyFile string) (*pem.Block, *pem.Block, error) {
	crt, err := ReadPem(crtFile)
	if err != nil {
		return nil, nil, err
	}

	key, err := ReadPem(keyFile)
	if err != nil {
		return nil, nil, err
	}

	return crt, key, nil
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
