package main

import (
	"flag"
	"log"

	"pypibot/rpc"
)

func main() {
	flagCrt := flag.String("crt", "kellegous.crt.pem", "")
	flagKey := flag.String("key", "kellegous.key.pem", "")
	flagCa := flag.String("ca", "data/ca.crt", "")
	flag.Parse()

	c, err := rpc.Dial("pypi.kellego.us:8081", *flagCa, *flagCrt, *flagKey)
	if err != nil {
		log.Panic(err)
	}

	log.Println(c)
}
