package main

import (
	"flag"
	"log"
	"net"
	"net/http"
)

func serveRpc(c net.Conn) {
}

func startRpc(addr string) error {
	a, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	s, err := net.ListenTCP("tcp", a)
	if err != nil {
		return err
	}

	go func() {
		for {
			c, err := s.Accept()
			if err != nil {
				log.Fatal(err)
			}

			go serveRpc(c)
		}
	}()

	return nil
}

func main() {
	flagHttpAddr := flag.String("http-addr", ":8080", "")
	flagRpcAddr := flag.String("rpc-addr", ":8081", "")
	flag.Parse()

	if err := startRpc(*flagRpcAddr); err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(*flagHttpAddr, nil))
}
