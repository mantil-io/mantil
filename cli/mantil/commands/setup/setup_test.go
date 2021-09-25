package setup

import (
	"testing"

	"github.com/mantil-io/mantil/aws"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	cli := aws.NewForTests(t)
	if cli == nil {
		t.Skip("skip: cli not initialized")
	}
	s := New(cli, Version{}, "")

	alreadyRun, err := s.isAlreadyRun()
	require.NoError(t, err)
	require.False(t, alreadyRun)
}
