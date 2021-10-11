package workspace

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
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

func TestFileStoreLoad(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	wfs, err := newWorkspacesFileStore(dir)

	ws, err := wfs.Load("pero")
	require.Error(t, err)
	require.Nil(t, ws)
	require.ErrorIs(t, err, ErrWorkspaceNotFound)

	ws, err = wfs.LoadOrNew("pero")
	require.NoError(t, err)
	require.NotNil(t, ws)
	require.Equal(t, ws.Name, "pero")
	require.NotEmpty(t, ws.UID)
	require.Len(t, ws.UID, 6)

	require.Empty(t, ws.Accounts)
}

func TestRandom(t *testing.T) {
	t.Skip()
	for i := 0; i < 10; i++ {
		j := rand.Int31()
		fmt.Printf("%s %x\n", uid(), j)
	}
}
