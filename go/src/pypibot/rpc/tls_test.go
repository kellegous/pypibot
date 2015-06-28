package rpc

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net"
	"testing"
	"time"
)

const (
	certCountry = "United States"
	certOrg     = "kellegous"
	certOrgUnit = "robotics"
)

func GenerateCert(bits int, caCrt, caKey *pem.Block) (*pem.Block, *pem.Block, error) {
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
		NotAfter:              time.Now().Add(500 * 24 * time.Hour),
		SerialNumber:          sn,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		Subject: pkix.Name{
			Country:            []string{certCountry},
			Organization:       []string{certOrg},
			OrganizationalUnit: []string{certOrgUnit},
		},
	}

	cac, err := x509.ParseCertificate(caCrt.Bytes)
	if err != nil {
		return nil, nil, err
	}

	cak, err := x509.ParsePKCS1PrivateKey(caKey.Bytes)
	if err != nil {
		return nil, nil, err
	}

	crt, err := x509.CreateCertificate(rand.Reader, tpl, cac, &prv.PublicKey, cak)
	if err != nil {
		return nil, nil, err
	}

	return &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: crt,
		}, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(prv),
		}, nil
}

func GenerateCa(bits int, hosts []string) (*pem.Block, *pem.Block, error) {
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
		NotAfter:              time.Now().Add(1000 * 24 * time.Hour),
		SerialNumber:          sn,
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		DNSNames:              hosts,
		Subject: pkix.Name{
			Country:            []string{certCountry},
			Organization:       []string{certOrg},
			OrganizationalUnit: []string{certOrgUnit},
		},
	}

	crt, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &prv.PublicKey, prv)
	if err != nil {
		return nil, nil, err
	}

	return &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: crt,
		}, &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(prv),
		}, nil
}

func newCaPool(crtPem *pem.Block) (*x509.CertPool, error) {
	p := x509.NewCertPool()

	crt, err := x509.ParseCertificate(crtPem.Bytes)
	if err != nil {
		return nil, err
	}

	p.AddCert(crt)

	return p, nil
}

func startServer(addr string, crtPem, keyPem *pem.Block) (io.Closer, error) {
	nl, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	prv, err := x509.ParsePKCS1PrivateKey(keyPem.Bytes)
	if err != nil {
		return nil, err
	}

	crt := tls.Certificate{
		Certificate: [][]byte{crtPem.Bytes},
		PrivateKey:  prv,
	}

	p, err := newCaPool(crtPem)
	if err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{crt},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    p,
	}

	tl := tls.NewListener(nl, cfg)

	go func() {
		for {
			c, err := tl.Accept()
			if err != nil {
				return
			}

			tc := c.(*tls.Conn)

			if err := tc.Handshake(); err != nil {
				log.Panic(err)
			}
		}
	}()

	return tl, nil
}

func connectClient(addr, host string, crtPem, keyPem, caCrtPem *pem.Block) error {
	prv, err := x509.ParsePKCS1PrivateKey(keyPem.Bytes)
	if err != nil {
		return err
	}

	crt := tls.Certificate{
		Certificate: [][]byte{crtPem.Bytes},
		PrivateKey:  prv,
	}

	p, err := newCaPool(caCrtPem)
	if err != nil {
		return err
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{crt},
		RootCAs:      p,
		ServerName:   host,
	}

	c, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		return err
	}
	defer c.Close()

	return nil
}

func TestIt(t *testing.T) {
	caCrtPem, caKeyPem, err := GenerateCa(2048, []string{"kellego.us"})
	if err != nil {
		t.Fatal(err)
	}

	crtPem, keyPem, err := GenerateCert(2048, caCrtPem, caKeyPem)
	if err != nil {
		t.Fatal(err)
	}

	srv, err := startServer(":9999", caCrtPem, caKeyPem)
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	if err := connectClient("localhost:9999", "kellego.us", crtPem, keyPem, caCrtPem); err != nil {
		t.Fatal(err)
	}
}
