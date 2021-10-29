package main

import (
	"fmt"
	"log"

	"github.com/mantil-io/mantil/event"
	"github.com/mantil-io/mantil/event/net"
	"github.com/mantil-io/mantil/kit/signal"
)

func main() {
	l, err := net.NewListener("./event-listener.creds")
	if err != nil {
		log.Fatal(err)
	}

	ch, err := l.Listen(signal.Interupt)
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	for buf := range ch {
		cc, err := event.NewCliCommand(buf)
		if err != nil {
			log.Fatal(err)
		}
		pretty, _ := cc.Pretty()
		i++
		fmt.Printf("%d\n%s\n\n", i, pretty)
	}
}
