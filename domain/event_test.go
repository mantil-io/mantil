package domain_test

import (
	"encoding/json"
	"testing"

	. "github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/gz"
	"github.com/stretchr/testify/require"
)

var testCliCommandJSON = `{
  "timestamp": 1637160510852,
  "duration": 5675,
  "version": "v0.1.22-38-g36beb6a",
  "args": [
    "mantil",
    "deploy"
  ],
  "device": {
    "os": "darwin",
    "arch": "amd64",
    "username": "ianic",
    "machineID": "5aa7f0e16feaaa6a0e609bb89d3662a0783e662eb049c1d5372f886895cb9136"
  },
  "workspace": {
    "name": "ianic",
    "nodes": 3,
    "projects": 1,
    "stages": 1,
    "functions": 2,
    "awsAccounts": 2,
    "awsRegions": 1
  },
  "project": {
    "name": "excuses",
    "stages": 1,
    "nodes": 1,
    "awsAccounts": 1,
    "awsRegions": 1
  },
  "stage": {
    "name": "dev",
    "node": "excuses",
    "functions": 2
  },
  "events": [
    {
      "timestamp": 1637160513108,
      "goBuild": {
        "name": "excuses",
        "duration": 308,
        "size": 11911830
      }
    },
    {
      "timestamp": 1637160513897,
      "goBuild": {
        "name": "ping",
        "duration": 756,
        "size": 11909380
      }
    },
    {
      "timestamp": 1637160516526,
      "deploy": {
        "functions": {
          "updated": 1
        },
        "buildDuration": 1064,
        "uploadDuration": 944,
        "uploadbytes": 6102455,
        "updateDuration": 1682
      }
    }
  ]
}`

func testCliCommand(t *testing.T) CliCommand {
	var cc CliCommand
	err := json.Unmarshal([]byte(testCliCommandJSON), &cc)
	require.NoError(t, err)
	return cc
}

func TestEventMarshal(t *testing.T) {
	cc := testCliCommand(t)
	buf, err := cc.Marshal()
	require.NoError(t, err)
	require.True(t, gz.IsZiped(buf))
	require.Len(t, buf, 253)

	bufUnziped, err := gz.Unzip(buf)
	require.NoError(t, err)
	require.Len(t, bufUnziped, 432)
}

func TestEventUnmarshal(t *testing.T) {
	tcc := testCliCommand(t)
	buf, err := tcc.Marshal()
	require.NoError(t, err)

	var cc CliCommand
	err = cc.Unmarshal(buf)
	require.NoError(t, err)
	require.Equal(t, testCliCommand(t), cc)
}

func TestEventUnmarshalUngzipped(t *testing.T) {
	tcc := testCliCommand(t)
	buf, err := tcc.Marshal()
	require.NoError(t, err)
	require.True(t, gz.IsZiped(buf))

	ungziped, err := gz.Unzip(buf)
	require.NoError(t, err)
	require.False(t, gz.IsZiped(ungziped))

	var cc CliCommand
	err = cc.Unmarshal(ungziped)
	require.NoError(t, err)
	require.Equal(t, testCliCommand(t), cc)
}
