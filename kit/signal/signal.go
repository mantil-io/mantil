package signal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

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

var Interupt context.Context

func init() {
	Interupt = interuptContext()
}
