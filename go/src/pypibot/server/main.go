package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"pypibot/store"
)

func doServe(args []string) {
	flags := flag.NewFlagSet("serve", flag.PanicOnError)
	flagDbPath := flags.String("dbpath", "data", "")
	flags.Parse(args)

	s, err := store.Open(*flagDbPath)
	if err != nil {
		log.Panic(err)
	}

	// var cfg config.Config

	// if err := cfg.ReadFromFile(*flagCfg); err != nil {
	// 	log.Panic(err)
	// }

	// if err := rpc.Serve(&cfg); err != nil {
	// 	log.Panic(err)
	// }

	// log.Panic(http.ListenAndServe(cfg.Web.Addr, nil))
	log.Panic(http.ListenAndServe(s.Config.Web.Addr, nil))
}

func doInit(args []string) {
	flags := flag.NewFlagSet("init", flag.PanicOnError)
	flagDbPath := flags.String("dbpath", "data", "")
	flags.Parse(args)

	if err := store.Create(*flagDbPath); err != nil {
		log.Panic(err)
	}

	fmt.Printf("Store created: %s\n", *flagDbPath)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s command [options] [args]\n", os.Args[0])
	os.Exit(1)
}

func main() {
	args := os.Args

	if len(args) < 2 {
		usage()
	}

	switch os.Args[1] {
	case "serve":
		doServe(args[1:])
	case "init":
		doInit(args[1:])
	default:
		usage()
	}
}
