package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"

	"pypibot/store"
)

// Client ...
type Client struct {
	c *tls.Conn
}

// Close ...
func (c *Client) Close() error {
	return c.c.Close()
}

// Ping ...
func (c *Client) Ping() (*PingRes, error) {
	var id int32 = 1

	if err := writeMsg(c.c, msgPingMsg, &PingReq{
		Id: id,
	}); err != nil {
		return nil, err
	}

	var res PingRes
	if err := readAndUnmarshalMsg(c.c, msgPingMsg, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// Dial ...
func Dial(addr string, srvCrtPem, crtPem, keyPem *pem.Block) (*Client, error) {
	prv, err := x509.ParsePKCS1PrivateKey(keyPem.Bytes)
	if err != nil {
		return nil, err
	}

	crt := tls.Certificate{
		Certificate: [][]byte{crtPem.Bytes},
		PrivateKey:  prv,
	}

	caCrt, err := x509.ParseCertificate(srvCrtPem.Bytes)
	if err != nil {
		return nil, err
	}

	p := x509.NewCertPool()
	p.AddCert(caCrt)

	cfg := &tls.Config{
		Certificates: []tls.Certificate{crt},
		RootCAs:      p,
		ServerName:   store.ServerName,
	}

	con, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		c: con,
	}, nil
}
