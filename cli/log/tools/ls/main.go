package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/signal"
	"github.com/nats-io/jsm.go"
	"github.com/nats-io/nats.go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

const (
	streamName   = "events"
	consumerName = "ls"
)

func run() error {
	nc, err := nats.Connect(nats.DefaultURL, nats.UseOldRequestStyle())
	if err != nil {
		return fmt.Errorf("Connect failed %w", err)
	}
	mgr, err := jsm.New(nc)
	if err != nil {
		return fmt.Errorf("jsm.New failed %w", err)
	}
	st, err := mgr.LoadStream(streamName)
	if err != nil {
		return fmt.Errorf("LoadStream failed %w", err)
	}
	cs, err := st.NewConsumer(jsm.DurableName(consumerName), jsm.DeliverAllAvailable())
	if err != nil {
		return fmt.Errorf("NewConsumer failed %w", err)
	}

	ctx := signal.Interupt
	for {
		nm, err := cs.NextMsgContext(ctx)
		if err != nil {
			if err == context.DeadlineExceeded {
				break
			}
			if err == context.Canceled {
				break
			}
			return err
		}
		cc, err := domain.NewCliCommand(nm.Data)
		if err != nil {
			log.Printf("Error %s", err)
			continue
			//return fmt.Errorf("NewCliCommand error %s", err)
		}

		streamSequence := 0
		//meta, _ := nm.JetStreamMetaData()
		if meta, err := jsm.ParseJSMsgMetadata(nm); err == nil {
			streamSequence = int(meta.StreamSequence())
		}
		pretty, _ := cc.Pretty()
		fmt.Printf("%d\n%s\n\n", streamSequence, pretty)

		nm.Ack()
	}
	return nil

}
