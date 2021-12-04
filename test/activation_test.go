package test

import (
	"testing"

	"github.com/mantil-io/mantil/kit/clitest"
)

func TestBeforeActivation(t *testing.T) {
	createNewWorkspaceWithoutToken(t)

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
