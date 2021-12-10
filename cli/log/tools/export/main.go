package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/signal"
	"github.com/nats-io/jsm.go"
	"github.com/nats-io/nats.go"
)

func main() {
	var dir string
	flag.StringVar(&dir, "dir", "/tmp", "dir where to create export file")
	flag.Parse()

	if err := run(dir); err != nil {
		log.Fatal(err)
	}
}

const (
	streamName   = "events"
	consumerName = "export"
)

func run(dir string) error {
	// open export file
	fileName := fmt.Sprintf("events-%s.log", time.Now().Format("2006-01-02-15-04-05"))
	filePath := filepath.Join(dir, fileName)
	log.Printf("writing to file %s", filePath)
	exportFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer exportFile.Close()

	// start nats connection
	nc, err := nats.Connect(nats.DefaultURL, nats.UseOldRequestStyle())
	if err != nil {
		return fmt.Errorf("Connect failed %w", err)
	}
	defer nc.Close()
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

	cnt := 0
	skipped := 0
	ctx := signal.Interupt
	for {
		cctx, cancel := context.WithTimeout(ctx, time.Second)
		nm, err := cs.NextMsgContext(cctx)
		if err != nil {
			if err == context.DeadlineExceeded {
				cancel()
				log.Printf("written %d events, skipped %d", cnt, skipped)
				if cnt == 0 {
					exportFile.Close()
					if err := os.Remove(filePath); err != nil {
						return err
					}
					log.Printf("deleted file %s", filePath)
				}
				return nil
			}
			if err == context.Canceled {
				cancel()
				break
			}
			cancel()
			return err
		}
		cancel()
		cc, err := domain.NewCliCommand(nm.Data)
		if err != nil {
			if err != nil {
				skipped++
				log.Printf("json error: %s", err)
				nm.Ack()
				continue
			}
			//return fmt.Errorf("NewCliCommand error %s", err)
		}

		buf, _ := cc.JSON()
		if _, err := fmt.Fprintln(exportFile, string(buf)); err != nil {
			return err
		}
		cnt++
		nm.Ack()
	}
	return nil
}
