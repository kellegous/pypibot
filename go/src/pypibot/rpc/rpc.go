package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"

	"github.com/golang/protobuf/proto"

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

	return nil, errors.New("certificate not authorized")
}

func readMsg(c net.Conn) (uint32, []byte, error) {
	var t uint32
	if err := binary.Read(c, binary.BigEndian, &t); err != nil {
		return 0, nil, err
	}

	var s uint32
	if err := binary.Read(c, binary.BigEndian, &s); err != nil {
		return 0, nil, err
	}

	b := make([]byte, int(s))
	if _, err := io.ReadFull(c, b); err != nil {
		return 0, nil, err
	}

	return t, b, nil
}

func writeMsg(c net.Conn, t uint32, m proto.Message) error {
	b, err := proto.Marshal(m)
	if err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, t); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, uint32(len(b))); err != nil {
		return err
	}

	if _, err := c.Write(b); err != nil {
		return err
	}

	return nil
}

func serve(c *tls.Conn, s *store.Store) {
	defer c.Close()

	if err := c.Handshake(); err != nil {
		log.Println(err)
		return
	}

	_, err := authenticate(c, s)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		t, m, err := readMsg(c)
		if err == io.EOF {
			return
		} else if err != nil {
			log.Panic(err)
			return
		}

		if err := dispatch(c, t, m); err != nil {
			log.Print(err)
			return
		}
	}
}

// Serve ...
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
