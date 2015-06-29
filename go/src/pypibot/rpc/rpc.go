package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"log"
	"net"
	"time"

	"pypibot/pb"
	"pypibot/store"
)

func newListener(s *store.Store) (net.Listener, error) {
	nl, err := net.Listen("tcp", s.Config.Rpc.Addr)
	if err != nil {
		return nil, err
	}

	cfg, err := s.ServerTlsConfig()
	if err != nil {
		return nil, err
	}

	return tls.NewListener(nl, cfg), nil
}

func authenticate(c *tls.Conn, s *store.Store) (*pb.User, error) {
	for _, cert := range c.ConnectionState().PeerCertificates {
		key, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
		if err != nil {
			continue
		}

		u, err := s.FindUser(key)
		if err != nil {
			continue
		}

		return u, nil
	}

	return nil, errors.New("certificate not authorized.")
}

func serve(c *tls.Conn, s *store.Store) {
	defer c.Close()

	if err := c.Handshake(); err != nil {
		log.Println(err)
		return
	}

	user, err := authenticate(c, s)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("user: %v\n", user)
	time.Sleep(10 * time.Second)
}

func Serve(s *store.Store) (io.Closer, error) {
	l, err := newListener(s)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				// the listener was closed
				return
			}

			go serve(c.(*tls.Conn), s)
		}
	}()

	return l, nil
}

type Client struct {
	c *tls.Conn
}

func (c *Client) Close() error {
	return c.c.Close()
}

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
