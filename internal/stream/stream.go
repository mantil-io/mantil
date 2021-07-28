package stream

import (
	"context"
	"fmt"
	"io"
	"log"

	mgo "github.com/atoz-technology/mantil.go"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

func createStream(subject string, in <-chan string, done chan interface{}) error {
	url := "connect.mantil.team"
	nc, err := nats.Connect(url, natsAuth())
	if err != nil {
		return err
	}
	go func() {
		for msg := range in {
			if err := nc.Publish(subject, []byte(msg)); err != nil {
				log.Printf("could not publish message - %v", err)
				continue
			}
		}
		if err := nc.Publish(subject, nil); err != nil {
			log.Printf("could not publish closing message - %v", err)
		}
		close(done)
	}()
	return nil
}

// copies log messages to ch
type logWriter struct {
	ch            chan string
	defaultWriter io.Writer
}

func newLogWriter() *logWriter {
	w := &logWriter{
		ch:            make(chan string),
		defaultWriter: log.Writer(),
	}
	log.SetOutput(w)
	return w
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	go func() {
		w.ch <- string(p)
	}()
	return w.defaultWriter.Write(p)
}

func (w *logWriter) close() {
	log.SetOutput(w.defaultWriter)
	close(w.ch)
}

func LogStream(subject string, callback func() error) error {
	w := newLogWriter()
	done := make(chan interface{})
	err := createStream(subject, w.ch, done)
	if err != nil {
		return err
	}
	callback()
	w.close()
	<-done
	return nil
}

func LambdaLogStream(ctx context.Context, callback func() error) error {
	var inbox string
	lctx, ok := mgo.FromContext(ctx)
	if !ok {
		return fmt.Errorf("error retrieving nats subject")
	}
	inbox = lctx.APIGatewayRequest.Headers["x-nats-inbox"]
	if inbox == "" {
		return fmt.Errorf("invalid nats subject")
	}
	return LogStream(inbox, callback)
}

func natsAuth() nats.Option {
	nkeySeed := "SUADEJU2KOHEGHFLBXJS4QF75E2II3PU63I3GCK4OBJLINOC7LDVEOX42A"
	nkeyUser := "UDQPHZBVNZCJSM5JXUDICFALEQ7Y5KPPAF7KGHTH77OGG7COQOJEYBZ7"

	opt := nats.Nkey(nkeyUser, func(nonce []byte) ([]byte, error) {
		user, err := nkeys.FromSeed([]byte(nkeySeed))
		if err != nil {
			return nil, err
		}
		return user.Sign(nonce)
	})
	return opt
}
