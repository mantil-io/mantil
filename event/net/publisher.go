package net

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/mantil-io/mantil/cli/build"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nkeys"
)

var (
	defaultNatsURL = "connect.ngs.global"
	subject        = "mantil.events"
)

type Publisher struct {
	subject string
	closed  chan struct{}
	nc      *nats.Conn
}

func NewPublisher() (*Publisher, error) {
	userJWT := build.EventPublisherCreds
	p := &Publisher{
		subject: subject,
	}
	if err := p.connect(userJWT); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Publisher) connect(userJWT string) error {
	closed := make(chan struct{})
	nc, err := nats.Connect(defaultNatsURL,
		//		nats.UserCredentials("/Users/ianic/.nkeys/creds/synadia/mantil/event-publisher.creds"),
		nats.UserJWT(
			func() (string, error) {
				return nkeys.ParseDecoratedJWT([]byte(userJWT))
			},
			func(nonce []byte) ([]byte, error) {
				kp, err := nkeys.ParseDecoratedNKey([]byte(userJWT))
				if err != nil {
					return nil, err
				}
				return kp.Sign(nonce)
			}),
		nats.ClosedHandler(func(_ *nats.Conn) {
			close(closed)
		}))

	if err != nil {
		return fmt.Errorf("connect error %w", err)
	}
	p.nc = nc
	p.closed = closed
	return nil
}

func (p *Publisher) Pub(payload []byte) error {
	ln := len(payload)
	lim := int(p.nc.MaxPayload()) - 100
	if ln <= lim {
		return p.nc.Publish(p.subject, payload)
	}
	for _, msg := range splitIntoMsgs(payload, p.subject, lim) {
		if err := p.nc.PublishMsg(msg); err != nil {
			return err
		}
	}
	return nil

}

// split payload into chunks and create nats.Msg with that chunk
// add correlationID to each chunk
// identify last chunk
func splitIntoMsgs(payload []byte, subject string, lim int) []*nats.Msg {
	chunks := split(payload, lim)
	correlationID := correlationID()
	var msgs []*nats.Msg
	for i, chunk := range chunks {
		msg := nats.NewMsg(subject)
		msg.Data = chunk
		msg.Header.Set(correlationIDHeaderKey, correlationID)
		if i == len(chunks)-1 { // last chunk indicator
			msg.Header.Set(lastChunkHeaderKey, lastChunkHeaderKey)
		}
		msgs = append(msgs, msg)
	}
	return msgs
}

var (
	correlationIDHeaderKey = "C"
	lastChunkHeaderKey     = "L"
)

func chunkID(nm *nats.Msg) string {
	if nm.Header == nil {
		return ""
	}
	return nm.Header.Get(correlationIDHeaderKey)
}

func isLastChunk(nm *nats.Msg) bool {
	if nm.Header == nil {
		return false
	}
	return nm.Header.Get(lastChunkHeaderKey) == lastChunkHeaderKey
}

// idea stolen from:  https://github.com/nats-io/nats-server/blob/fd9e9480dad9498ed8109e659fc8ed5c9b2a1b41/server/nkey.go#L41
func correlationID() string {
	var rndData [4]byte
	data := rndData[:]
	_, _ = io.ReadFull(rand.Reader, data)
	var encoded [6]byte
	base64.RawURLEncoding.Encode(encoded[:], data)
	return string(encoded[:])
}

func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf)
	}
	return chunks
}

func (p *Publisher) Close() error {
	if err := p.nc.Drain(); err != nil {
		return err
	}
	if p.closed != nil {
		<-p.closed
	}
	return nil
}
