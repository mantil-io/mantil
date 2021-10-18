package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStore(t *testing.T) *FileStore {
	fs := &FileStore{
		workspaceFile: "testdata/workspace.yml",
		projectRoot:   "testdata",
	}
	err := fs.restore()
	require.NoError(t, err)
	return fs
}

func TestLoad(t *testing.T) {
	fs := testStore(t)

	require.NotNil(t, fs.workspace)
	assert.Equal(t, "my-workspace", fs.workspace.Name)

	// test that link exists
	require.NotNil(t, fs.project)
	require.Equal(t, fs.project.workspace, fs.workspace)
	require.Len(t, fs.project.Stages, 1)
	stage := fs.project.Stages[0]
	require.Equal(t, fs.project, stage.project)
	require.Equal(t, fs.workspace.Accounts[0], stage.account)
	require.Equal(t, fs.workspace.Accounts[0].workspace, fs.workspace)
}

func TestStageResourceNaming(t *testing.T) {
	fs := testStore(t)

	stage := fs.DefaultStage()

	require.Equal(t, "stages/misteriozo/mister1", stage.BucketPrefix())
	require.Equal(t, "misteriozo-mister1", stage.LogGroupsPrefix())
	require.Equal(t, "misteriozo-mister1-ping-fpdtuji", stage.Functions[0].LambdaName())
}

func TestAccountResourceNaming(t *testing.T) {
	fs := testStore(t)
	ac := fs.Workspace().Account("dev")

	require.Equal(t, "mantil-setup-fpdtuji", ac.SetupStackName())
	require.Equal(t, "mantil-setup-fpdtuji", ac.SetupLambdaName())
}
