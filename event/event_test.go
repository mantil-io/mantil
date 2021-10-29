package event_test

import (
	"testing"

	. "github.com/mantil-io/mantil/event"
	"github.com/stretchr/testify/require"
)

var testCliCommand = CliCommand{
	Timestamp: 1234567890,
	Version:   "v1.2.3",
	Command:   "mantil aws install",
	Args:      []string{"pero", "zdero"},
	Workspace: "my-workspace",
	Project:   "my-project",
	Stage:     "my-stage",
	Events: []Event{
		{
			Deploy: &Deploy{BuildDuration: 1, UploadDuration: 2, UpdateDuration: 3, UploadMiB: 4},
		},
	},
}

func TestEventMarshal(t *testing.T) {
	buf, err := testCliCommand.Marshal()
	require.NoError(t, err)
	require.True(t, IsGziped(buf))
	require.Len(t, buf, 150)

	bufUnziped, err := Gunzip(buf)
	require.NoError(t, err)
	require.Len(t, bufUnziped, 164)

	expected := `{"t":1234567890,"v":"v1.2.3","c":"mantil aws install","a":["pero","zdero"],"w":"my-workspace","p":"my-project","s":"my-stage","e":[{"d":{"b":1,"u":2,"m":4,"d":3}}]}`
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
	require.True(t, IsGziped(buf))

	ungziped, err := Gunzip(buf)
	require.NoError(t, err)
	require.False(t, IsGziped(ungziped))

	var cc CliCommand
	err = cc.Unmarshal(ungziped)
	require.NoError(t, err)
	require.Equal(t, testCliCommand, cc)
}
