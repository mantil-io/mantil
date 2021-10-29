package net

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func TestListenerJoinChunks(t *testing.T) {
	// create chunks
	msgs1 := splitIntoMsgs([]byte(testSplitData), "subject", 16)
	msgs2 := splitIntoMsgs([]byte(testSplitData), "subject", 7)
	require.Len(t, msgs1, 8)
	require.Len(t, msgs2, 18)

	// init listener
	l := &Listener{
		chunks: make(map[string][][]byte),
	}
	nmsgs := make(chan *nats.Msg, len(msgs1)+len(msgs2)) // enough room for all messages
	out := make(chan []byte, 0)                          // len=0, block on write

	ctx, cancel := context.WithCancel(context.Background())
	go l.loop(ctx, nmsgs, out)

	// interleave chunks
	for i := 0; i < len(msgs2); i++ {
		nmsgs <- msgs2[i]
		if i < len(msgs1) {
			nmsgs <- msgs1[i]
		}
	}

	out1 := <-out
	out2 := <-out

	require.Equal(t, string(out1), testSplitData)
	require.Equal(t, string(out2), testSplitData)
	require.Len(t, l.chunks, 0)
	cancel()
}
