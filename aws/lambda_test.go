package aws

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLambdaExists(t *testing.T) {
	cli := NewForTests(t)
	if cli == nil {
		t.Skip("skip: cli not initialized")
	}
	found, err := cli.LambdaExists("this-function-dont-exists")
	require.NoError(t, err)
	require.False(t, found)
}
