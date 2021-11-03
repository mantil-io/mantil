package main

import (
	"fmt"
	"log"

	"github.com/mantil-io/mantil/cli/log/net"
	"github.com/mantil-io/mantil/domain"
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
		cc, err := domain.NewCliCommand(buf)
		if err != nil {
			log.Fatal(err)
		}
		pretty, _ := cc.Pretty()
		i++
		fmt.Printf("%d\n%s\n\n", i, pretty)
	}
}
