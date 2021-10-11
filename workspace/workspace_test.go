package workspace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWorkspaceAccount(t *testing.T) {
	var w Workspace
	require.Len(t, w.Accounts, 0)

	t.Run("add new account", func(t *testing.T) {
		a, err := w.NewAccount("first", "accountID", "region", "bucket", "path")
		require.NoError(t, err)
		require.Equal(t, a.ID, "accountID")
		require.Equal(t, a.Region, "region")
		require.Equal(t, a.Functions.Bucket, "bucket")
		require.Equal(t, a.Functions.Path, "path")
		require.NotEmpty(t, a.Keys.Public)
		require.NotEmpty(t, a.Keys.Private)
		require.Len(t, w.Accounts, 1)
	})

	t.Run("add existing account", func(t *testing.T) {
		require.True(t, w.accountExists("first"))
		a, err := w.NewAccount("first", "accountID", "region", "bucket", "path")
		require.Nil(t, a)
		require.ErrorIs(t, err, ErrAccountExists)
	})
}
