package net

import (
	"bytes"
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

func NewListener(userJWTorCredsFile string) (*Listener, error) {
	nc, err := natsConnect(userJWTorCredsFile)
	if err != nil {
		return nil, fmt.Errorf("connect error %w", err)
	}
	return &Listener{
		chunks: make(map[string][][]byte),
		nc:     nc,
	}, nil
}

type Listener struct {
	chunks map[string][][]byte
	nc     *nats.Conn
}

func (l *Listener) Listen(ctx context.Context) (chan []byte, error) {
	nmsgs := make(chan *nats.Msg, 1024)
	sub, err := l.nc.ChanSubscribe(subject, nmsgs)
	if err != nil {
		return nil, fmt.Errorf("subscribe error %w", err)
	}
	out := make(chan []byte)
	go func() {
		defer close(out)
		defer sub.Unsubscribe()
		l.loop(ctx, nmsgs, out)
	}()
	return out, nil
}

func (l *Listener) loop(ctx context.Context, nmsgs chan *nats.Msg, out chan []byte) {
	for {
		select {
		case nm := <-nmsgs:
			id := chunkID(nm)
			if id == "" {
				out <- nm.Data
				break
			}
			l.pushChunk(id, nm.Data)
			if isLastChunk(nm) {
				out <- l.popChunk(id)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (l *Listener) pushChunk(id string, payload []byte) {
	l.chunks[id] = append(l.chunks[id], payload)
}

func (l *Listener) popChunk(id string) []byte {
	payload := bytes.Join(l.chunks[id], []byte{})
	delete(l.chunks, id)
	return payload
}
