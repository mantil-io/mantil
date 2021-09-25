package setup

import (
	"testing"

	"github.com/mantil-io/mantil/aws"
	"github.com/stretchr/testify/require"
)

func TestCreateLambda(t *testing.T) {
	cli := aws.NewForTests(t)
	if cli == nil {
		t.Skip("skip: cli not initialized")
	}
	cmd := New(cli, Version{}, "")
	// empty at start
	alreadyRun, err := cmd.isAlreadyRun()
	require.NoError(t, err)
	require.False(t, alreadyRun)
	// create lambda
	err = cmd.ensureLambdaExists()
	require.NoError(t, err)
	// check exists
	exists, err := cmd.awsClient.LambdaExists(lambdaName)
	require.NoError(t, err)
	require.True(t, exists)
	// and one more
	alreadyRun, err = cmd.isAlreadyRun()
	require.NoError(t, err)
	require.True(t, alreadyRun)
	// clanup
	err = cmd.deleteLambda()
	require.NoError(t, err)
	// check we are at clean
	alreadyRun, err = cmd.isAlreadyRun()
	require.NoError(t, err)
	require.False(t, alreadyRun)
}
