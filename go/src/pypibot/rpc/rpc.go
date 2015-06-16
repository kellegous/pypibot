package rpc

import (
	"crypto/tls"
	"log"
	"net"

	"pypibot/store"
)

func newListener(s *store.Store) (net.Listener, error) {
	crtFile, keyFile := s.RpcCertFiles()

	crt, err := tls.LoadX509KeyPair(crtFile, keyFile)
	if err != nil {
		return nil, err
	}

	l, err := net.Listen("tcp", s.Config.Rpc.Addr)
	if err != nil {
		return nil, err
	}

	return tls.NewListener(l, &tls.Config{
		Certificates: []tls.Certificate{crt},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}), nil
}

func serve(c *tls.Conn, s *store.Store) {
	defer c.Close()

	if err := c.Handshake(); err != nil {
		log.Println(err)
		return
	}

	// TODO(knorton): Authenticate and dispatch by client type.
	log.Printf("%s connected.", c.RemoteAddr())
}

func Serve(s *store.Store) error {
	l, err := newListener(s)
	if err != nil {
		return err
	}

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}

			go serve(c.(*tls.Conn), s)
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
