package main

import (
	"log"

	"pypibot/rpc"
)

func main() {
	c, err := rpc.Dial("localhost:8081")
	if err != nil {
		log.Panic(err)
	}

	log.Println(c)
}
