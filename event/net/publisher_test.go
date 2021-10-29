package net

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCorrelationID(t *testing.T) {
	for i := 0; i < 100; i++ {
		c := correlationID()
		require.Len(t, c, 6)
		//t.Logf("correlationID: %s", c)
	}
}

var (
	testSplitData = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
)

func TestSplit(t *testing.T) {
	chunks := split([]byte(testSplitData), 16)
	require.Len(t, chunks, 8)
	for i := 0; i < 7; i++ {
		require.Len(t, chunks[i], 16)
		//t.Logf("chunk %d %s", i, chunks[i])
	}
	require.Len(t, chunks[7], 11)
	require.Equal(t, string(bytes.Join(chunks, []byte{})), testSplitData)
}

func TestSplitIntoMsgs(t *testing.T) {
	msgs := splitIntoMsgs([]byte(testSplitData), "subject", 16)

	// test len of the payload
	require.Len(t, msgs, 8)
	for i := 0; i < 7; i++ {
		require.Len(t, msgs[i].Data, 16)
	}
	require.Len(t, msgs[7].Data, 11)

	// all messages have same correlationID header seet
	correlationID := msgs[0].Header.Get(correlationIDHeaderKey)
	for i := 0; i < len(msgs); i++ {
		require.Equal(t, msgs[i].Header.Get(correlationIDHeaderKey), correlationID)
		_, found := msgs[i].Header[correlationIDHeaderKey]
		require.True(t, found)
	}

	// last message has last chunk header set
	for i := 0; i < 7; i++ {
		require.Equal(t, msgs[i].Header.Get(lastChunkHeaderKey), "")
		_, found := msgs[i].Header[lastChunkHeaderKey]
		require.False(t, found)
	}
	_, found := msgs[7].Header[lastChunkHeaderKey]
	require.True(t, found)
	require.Equal(t, msgs[7].Header.Get(lastChunkHeaderKey), lastChunkHeaderKey)
}
