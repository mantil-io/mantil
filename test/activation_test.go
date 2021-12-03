package test

import (
	"io/ioutil"
	"testing"

	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/clitest"
	"github.com/stretchr/testify/require"
)

func TestBeforeActivation(t *testing.T) {
	workspacePath, err := ioutil.TempDir("/tmp", "mantil-workspace-")
	require.NoError(t, err)
	t.Setenv(domain.EnvWorkspacePath, workspacePath)
	t.Logf("workspace path: %s", workspacePath)

	r := clitest.New(t)
	r.Run("mantil", "--help").Success()
	r.Run("mantil", "aws", "install", "--help").Success()

	r.Run("mantil", "new", "foo").Fail().
		Stdout().Contains("Mantil is not activated")

	r.Run("mantil", "test").Fail().
		Stdout().Contains("Mantil is not activated")

	r.Run("mantil", "deploy").Fail().
		Stdout().Contains("Mantil is not activated")

}
