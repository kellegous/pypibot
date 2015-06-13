package main

import (
	"flag"
	"log"
	"net/http"

	"pypibot/config"
	"pypibot/rpc"
)

func main() {
	flagCfg := flag.String("cfg", "config.json", "")
	flag.Parse()

	var cfg config.Config

	if err := cfg.ReadFromFile(*flagCfg); err != nil {
		log.Panic(err)
	}

	if err := rpc.Serve(&cfg); err != nil {
		log.Panic(err)
	}

	log.Panic(http.ListenAndServe(cfg.Web.Addr, nil))
}
