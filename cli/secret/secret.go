// Package secrets provides credentials embeded into cli binary. Currently that
// are ngs nats access credentilas. All of them have only necessary rights to
// publish/subscribe to selected subject. They are hardly usefull for anyone who
// would potentialy extract them from binary.
package secret

import _ "embed"

// For publishing mantil events.
// Created with:
// nsc add user -n event-publisher --allow-pub mantil.events --deny-sub '*'
//go:embed event-publisher.creds
var EventPublisherCreds string

// Publisher and subscriber of log from lambda function execution. Listener
// creates subject and sends it to the publisher. Publisher can only publish to
// inboxes.
// Created with:
// nsc add user -n logs-publisher --allow-pub '_INBOX.>' --deny-sub '*'
//go:embed logs-publisher.creds
var LogsPublisherCreds string

// Created with:
// nsc add user -n logs-listener --allow-sub '_INBOX.>' --deny-pub '*'
//go:embed logs-listener.creds
var LogsListenerCreds string
