package main

import (
	"log"
	"time"

	"github.com/mantil-io/mantil/event"
	"github.com/mantil-io/mantil/event/net"
)

func main() {
	p, err := net.NewPublisher()
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

var testCliCommand = event.CliCommand{
	Timestamp: time.Now().UnixNano(),
	Version:   "v1.2.3",
	Command:   "mantil aws install 1",
	Args:      []string{"pero", "zdero"},
	Workspace: "my-workspace",
	Project:   "my-project",
	Stage:     "my-stage",
	Events: []event.Event{
		{
			Deploy: &event.Deploy{BuildDuration: 1, UploadDuration: 2, UpdateDuration: 3, UploadMiB: 4},
		},
	},
}

func largeTestCommnad() event.CliCommand {
	cc := testCliCommand
	var events []event.Event
	for i := int64(1); i < 1000; i++ {
		d := event.Deploy{BuildDuration: i + 1, UploadDuration: i + 2, UpdateDuration: i + 3, UploadMiB: i + 4}
		events = append(events, event.Event{Deploy: &d})
	}
	cc.Events = events
	return cc
}
