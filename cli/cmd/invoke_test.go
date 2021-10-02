package cmd

import (
	"testing"

	"github.com/mantil-io/mantil/cli/cmd/project"

	"github.com/stretchr/testify/require"
)

func TestInvokeURL(t *testing.T) {
	cmd := &invokeCmd{
		ctx: &project.Context{},
	}

	err := cmd.invoke()
	// stage rest endpoint is not set
	require.Error(t, err)
}
