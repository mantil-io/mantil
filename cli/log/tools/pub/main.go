package main

import (
	_ "embed"
	"log"
	"time"

	"github.com/mantil-io/mantil/cli/log/net"
	"github.com/mantil-io/mantil/domain"
)

//go:embed event-publisher.creds
var EventPublisherCreds string

func main() {
	p, err := net.NewPublisher(EventPublisherCreds)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()
	cc := testCliCommand
	//cc := largeTestCommnad()
	buf, err := cc.Marshal()
	if err != nil {
		log.Fatal(err)
	}
	if err := p.Pub(buf); err != nil {
		log.Fatal(err)
	}
}

var testCliCommand = domain.CliCommand{
	Timestamp: time.Now().UnixNano(),
	Version:   "v1.2.3",
	//Command:   "mantil aws install 1",
	Args: []string{"pero", "zdero"},
	//Workspace.Name: "my-workspace",
	//Project:        "my-project",
	//Stage:          "my-stage",
	Events: []domain.Event{
		{
			Deploy: &domain.Deploy{BuildDuration: 1, UploadDuration: 2, UpdateDuration: 3, UploadBytes: 4},
		},
	},
}

func largeTestCommnad() domain.CliCommand {
	cc := testCliCommand
	var events []domain.Event
	for i := 1; i < 1000; i++ {
		d := domain.Deploy{BuildDuration: i + 1, UploadDuration: i + 2, UpdateDuration: i + 3, UploadBytes: i + 4}
		events = append(events, domain.Event{Deploy: &d})
	}
	cc.Events = events
	return cc
}
