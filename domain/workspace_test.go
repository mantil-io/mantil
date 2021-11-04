package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWorkspaceNode(t *testing.T) {
	var w Workspace
	require.Len(t, w.Nodes, 0)

	t.Run("add new node", func(t *testing.T) {
		a, err := w.NewNode("first", "accountID", "region", "bucket", "path")
		require.NoError(t, err)
		require.Equal(t, a.ID, "accountID")
		require.Equal(t, a.Region, "region")
		require.Equal(t, a.Functions.Bucket, "bucket")
		require.Equal(t, a.Functions.Path, "path")
		require.NotEmpty(t, a.Keys.Public)
		require.NotEmpty(t, a.Keys.Private)
		require.NotEmpty(t, a.UID)
		require.Len(t, w.Nodes, 1)
	})

	t.Run("add existing node", func(t *testing.T) {
		require.True(t, w.nodeExists("first"))
		a, err := w.NewNode("first", "accountID", "region", "bucket", "path")
		require.Nil(t, a)
		var ea *NodeExistsError
		require.ErrorAs(t, err, &ea)
		require.Equal(t, "first", ea.Name)
	})
}
