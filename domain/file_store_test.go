package domain

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
	require.Equal(t, fs.workspace.Nodes[0], stage.node)
	require.Equal(t, fs.workspace.Nodes[0].workspace, fs.workspace)
}

func TestSave(t *testing.T) {
	fs := testStore(t)
	err := fs.Store()
	require.NoError(t, err)
}

func TestStageResourceNaming(t *testing.T) {
	fs := testStore(t)

	stage := fs.DefaultStage()

	require.Equal(t, "functions/misteriozo/mister1", stage.FunctionsBucketPrefix())
	require.Equal(t, "state/misteriozo/mister1", stage.StateBucketPrefix())
	require.Equal(t, "misteriozo-mister1", stage.LogGroupsPrefix())
	require.Equal(t, "misteriozo-mister1-ping-fpdtuji", stage.Functions[0].LambdaName())
	require.Equal(t, "misteriozo-mister1-%s-fpdtuji", stage.ResourceNamingTemplate())
}

func TestStageResourceTags(t *testing.T) {
	fs := testStore(t)

	stage := fs.DefaultStage()
	tags := stage.ResourceTags()
	require.NotEmpty(t, tags)

	assert.Equal(t, "my-workspace", tags[TagWorkspace])
	assert.Equal(t, "fpdtuji", tags[TagKey])
	assert.Equal(t, "misteriozo", tags[TagProjectName])
	assert.Equal(t, "mister1", tags[TagStageName])
}

func TestNodeResourceNaming(t *testing.T) {
	fs := testStore(t)
	n := fs.Workspace().Node("dev")

	require.Equal(t, "mantil-setup-fpdtuji", n.SetupStackName())
	require.Equal(t, "mantil-setup-fpdtuji", n.SetupLambdaName())
}
