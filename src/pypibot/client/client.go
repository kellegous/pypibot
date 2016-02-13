package main

import (
	"flag"
	"log"
	"time"

	"pypibot/auth"
	"pypibot/rpc"
)

func main() {
	flagCrt := flag.String("crt", "crt.pem", "")
	flagKey := flag.String("key", "key.pem", "")
	flagSrvCrt := flag.String("srvCrt", "data/srv.crt.pem", "")
	flag.Parse()

	crtPem, keyPem, err := auth.ReadBothPems(*flagCrt, *flagKey)
	if err != nil {
		log.Panic(err)
	}

	srvCrtPem, err := auth.ReadPem(*flagSrvCrt)
	if err != nil {
		log.Panic(err)
	}

	clt, err := rpc.Dial("pypi.kellego.us:8081", srvCrtPem, crtPem, keyPem)
	if err != nil {
		log.Panic(err)
	}

	for {
		res, err := clt.Ping()
		if err != nil {
			log.Panic(err)
		}
		log.Println(res)

		time.Sleep(2 * time.Second)
	}
}
