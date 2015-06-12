package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/jordanorelli/moon"
)

type Config struct {
	WebAddr string `name: web_addr; required: true`
	RpcAddr string `name: rpc_addr; required: true`
}

func (c *Config) readFrom(filename string) error {
	doc, err := moon.ReadFile(filename)
	if err != nil {
		return err
	}

	return doc.Fill(c)
}

func main() {
	flagCfg := flag.String("cfg", "config.moon", "")
	flag.Parse()

	var cfg Config

	if err := cfg.readFrom(*flagCfg); err != nil {
		log.Panic(err)
	}

	fmt.Println(cfg)
}
