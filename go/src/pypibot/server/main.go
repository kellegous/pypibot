package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"pypibot/api"
	"pypibot/auth"
	"pypibot/pb"
	"pypibot/rpc"
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

	if err := rpc.Serve(s); err != nil {
		log.Panic(err)
	}

	r := http.NewServeMux()

	api.Install(r, s)

	log.Panic(http.ListenAndServe(s.Config.Web.Addr, r))
}

func doInitStore(args []string) {
	flags := flag.NewFlagSet("init-store", flag.PanicOnError)
	flagDbPath := flags.String("dbpath", "data", "")
	flags.Parse(args)

	if err := store.Create(*flagDbPath); err != nil {
		log.Panic(err)
	}

	fmt.Printf("Store created: %s\n", *flagDbPath)
}

func doAddUser(args []string) {
	flags := flag.NewFlagSet("add-user", flag.PanicOnError)
	flagDbPath := flags.String("dbpath", "data", "")
	flags.Parse(args)

	if flags.NArg() != 4 {
		fmt.Fprintf(os.Stderr, "usage %s %s email name crt key\n", os.Args[0], args[0])
		os.Exit(1)
	}

	s, err := store.Open(*flagDbPath)
	if err != nil {
		log.Panic(err)
	}

	caCrt, caKey := s.CaCertFiles()

	if _, err := auth.GenerateCert(
		"kellego.us",
		caCrt,
		caKey,
		flags.Arg(2),
		flags.Arg(3)); err != nil {
		log.Panic(err)
	}

	e := flags.Arg(0)
	n := flags.Arg(1)
	t := pb.User_PERSON

	if err := s.AddUserWithKeyFromFile(&pb.User{
		Email: &e,
		Name:  &n,
		Type:  &t,
	}, flags.Arg(3)); err != nil {
		log.Panic(err)
	}
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
		doServe(args[2:])
	case "init-store":
		doInitStore(args[2:])
	case "add-user":
		doAddUser(args[2:])
	default:
		usage()
	}
}
