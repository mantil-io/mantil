package domain_test

import (
	"testing"

	. "github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/gz"
	"github.com/stretchr/testify/require"
)

var testCliCommand = CliCommand{
	Timestamp: 1234567890,
	Version:   "v1.2.3",
	Command:   "mantil aws install",
	Args:      []string{"pero", "zdero"},
	Events: []Event{
		{
			Deploy: &Deploy{BuildDuration: 1, UploadDuration: 2, UpdateDuration: 3, UploadBytes: 4},
		},
	},
}

func TestEventMarshal(t *testing.T) {
	buf, err := testCliCommand.Marshal()
	require.NoError(t, err)
	require.True(t, gz.IsZiped(buf))
	require.Len(t, buf, 135)

	bufUnziped, err := gz.Unzip(buf)
	require.NoError(t, err)
	require.Len(t, bufUnziped, 155)

	expected := `{"t":1234567890,"v":"v1.2.3","c":"mantil aws install","a":["pero","zdero"],"m":{},"w":{},"p":{},"s":{},"e":[{"d":{"f":{},"s":{},"b":1,"u":2,"m":4,"d":3}}]}`
	require.Equal(t, string(bufUnziped), expected)
}

func TestEventUnmarshal(t *testing.T) {
	buf, err := testCliCommand.Marshal()
	require.NoError(t, err)

	var cc CliCommand
	err = cc.Unmarshal(buf)
	require.NoError(t, err)
	require.Equal(t, testCliCommand, cc)
}

func TestEventUnmarshalUngzipped(t *testing.T) {
	buf, err := testCliCommand.Marshal()
	require.NoError(t, err)
	require.True(t, gz.IsZiped(buf))

	ungziped, err := gz.Unzip(buf)
	require.NoError(t, err)
	require.False(t, gz.IsZiped(ungziped))

	var cc CliCommand
	err = cc.Unmarshal(ungziped)
	require.NoError(t, err)
	require.Equal(t, testCliCommand, cc)
}
