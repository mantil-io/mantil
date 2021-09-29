package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/mantil-io/mantil/aws"
	"github.com/stretchr/testify/require"
)

func TestLogsFormatEvent(t *testing.T) {
	cmd := &logsCmd{}

	tv, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", "2021-09-28 11:43:51.015 +0200 CEST")
	require.NoError(t, err)

	le := aws.LogEvent{
		Timestamp: cmd.timestamp(tv),
		Message:   "message",
	}

	ets := cmd.eventTs(le)
	require.Equal(t, tv, ets)

	fe := fmt.Sprintf("%v message", tv)
	require.Equal(t, fe, cmd.formatEvent(le))
}
