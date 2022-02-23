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
	assert.Equal(t, "my-workspace-id", fs.workspace.ID)

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

func TestNodeResourceNaming(t *testing.T) {
	fs := testStore(t)
	n := fs.Workspace().Node("dev")

	require.Equal(t, "mantil-setup-fpdtuji", n.SetupStackName())
	require.Equal(t, "mantil-setup-fpdtuji", n.SetupLambdaName())
}

func TestFileStoreDeveloperStore(t *testing.T) {
	ds, err := NewSingleDeveloperProjectStore()
	require.Error(t, err)
	require.Nil(t, ds)

	ds, err = NewSingleDeveloperWorkspaceStore()
	require.NoError(t, err)
	require.NotNil(t, ds)
}

func TestFileStoreResources(t *testing.T) {
	fs := testStore(t)
	require.NotNil(t, fs.Project())
	require.NotNil(t, fs.DefaultStage())
	require.NotEmpty(t, fs.projectRoot)
	require.NotNil(t, fs.Stage(""))
	require.NotNil(t, fs.Stage("mister1"))
	require.Nil(t, fs.Stage("non-existent"))
}

func TestStoreAsCliWorkspace(t *testing.T) {
	fs := testStore(t)

	cw := fs.AsCliWorkspace()
	require.NotNil(t, cw)
	require.Equal(t, fs.workspace.ID, cw.ID)
	require.Equal(t, 1, cw.Nodes)
	require.Equal(t, 1, cw.Projects)
	require.Equal(t, 1, cw.Stages)
	require.Equal(t, 4, cw.Functions)
}

func TestStoreNewProject(t *testing.T) {
	fs := testStore(t)

	err := fs.NewProject("project", "/tmp/project")
	require.NoError(t, err)
}
