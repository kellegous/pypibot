package rpc

import (
	"crypto/tls"
	"log"
	"net"

	"pypibot/config"
)

func newListener(cfg *config.Config) (net.Listener, error) {
	crt, err := tls.LoadX509KeyPair(cfg.Rpc.Crt, cfg.Rpc.Key)
	if err != nil {
		return nil, err
	}

	l, err := net.Listen("tcp", cfg.Rpc.Addr)
	if err != nil {
		return nil, err
	}

	return tls.NewListener(l, &tls.Config{
		Certificates: []tls.Certificate{crt},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}), nil
}

func serve(c *tls.Conn, cfg *config.Config) {
	defer c.Close()

	if err := c.Handshake(); err != nil {
		log.Println(err)
		return
	}

	log.Printf("len(certs) = %d", len(c.ConnectionState().PeerCertificates))
	for _, crt := range c.ConnectionState().PeerCertificates {
		log.Println(crt)
	}

	// TODO(knorton): Authenticate and dispatch by client type.
	log.Printf("%s connected.", c.RemoteAddr())
}

func Serve(cfg *config.Config) error {
	l, err := newListener(cfg)
	if err != nil {
		return err
	}

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}

			go serve(c.(*tls.Conn), cfg)
		}
	}()

	return nil
}

type Client struct {
	c *tls.Conn
}

func Dial(addr string) (*Client, error) {
	c, err := tls.Dial("tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		c: c,
	}, nil
}
