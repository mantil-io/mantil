package main

import (
	"log"

	"github.com/mantil-io/mantil/event/net"
)

func main() {
	p, err := net.NewPublisher()
	if err != nil {
		log.Fatal(err)
	}

}
