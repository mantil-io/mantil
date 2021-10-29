package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mantil-io/mantil/event"
	"github.com/mantil-io/mantil/event/net"
)

func main() {
	l, err := net.NewListener("./event-listener.creds")
	if err != nil {
		log.Fatal(err)
	}
	ctx := interuptContext()
	ch, err := l.Listen(ctx)
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

func waitForInterupt() {
	c := make(chan os.Signal, 1)
	//SIGINT je ctrl-C u shell-u, SIGTERM salje upstart kada se napravi sudo stop ...
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

// InteruptContext returns context which will be closed on application interupt
func interuptContext() context.Context {
	ctx, stop := context.WithCancel(context.Background())
	go func() {
		waitForInterupt()
		stop()
	}()
	return ctx
}
