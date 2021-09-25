package aws

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const testsProfileEnv = "MANTIL_TESTS_AWS_PROFILE"

func NewForTests(t *testing.T) *AWS {
	val, ok := os.LookupEnv(testsProfileEnv)
	if !ok {
		t.Logf("environment vairable %s not found", testsProfileEnv)
		return nil
	}
	cli, err := NewFromProfile(val)
	if err != nil {
		t.Fatal(err)
	}
	return cli
}

func TestLambdaExists(t *testing.T) {
	cli := NewForTests(t)
	if cli == nil {
		t.Skip("skip: cli not initialized")
	}
	found, err := cli.LambdaExists("this-function-dont-exists")
	require.NoError(t, err)
	require.False(t, found)
}
