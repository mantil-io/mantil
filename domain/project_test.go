package domain_test

import (
	"testing"

	. "github.com/mantil-io/mantil/domain"
	"github.com/stretchr/testify/require"
)

func TestProjectGetStage(t *testing.T) {
	project := testProject(t)

	stage := project.Stage("my-stage")

	require.NotNil(t, stage)
	require.Equal(t, "my-stage", stage.Name)
}

func TestProjectDefaultStage(t *testing.T) {
	project := testProject(t)
	ds := project.DefaultStage()

	// if there is only one stage that is default
	require.NotNil(t, ds)
	require.Equal(t, "my-stage", ds.Name)

	_, err := project.NewStage("stage2", "node1", "path")
	require.NoError(t, err)

	ds = project.DefaultStage()
	require.NotNil(t, ds)
	require.Equal(t, "my-stage", ds.Name)
}

func TestProjectNewStage(t *testing.T) {
	project := testProject(t)
	require.Len(t, project.Stages, 1)

	s2, err := project.NewStage("stage2", "node1", "path")
	require.NoError(t, err)

	require.Len(t, project.Stages, 2)
	require.Equal(t, "node1", s2.Node().Name)
	require.Equal(t, project, s2.Project())
	require.False(t, s2.Default)

	_, err = project.NewStage("very-long-name-that-fails-validation", "node1", "path")
	require.Error(t, err)
}

func TestProjectFirstNewStageIsDefault(t *testing.T) {
	project := testProject(t)
	require.Len(t, project.Stages, 1)
	project.RemoveStage("my-stage")
	require.Len(t, project.Stages, 0)

	s2, err := project.NewStage("stage2", "node1", "path")
	require.NoError(t, err)

	require.Equal(t, "node1", s2.Node().Name)
	require.Equal(t, project, s2.Project())
	require.True(t, s2.Default)
}

func TestProjectNewStageWithEmptyName(t *testing.T) {
	project := testProject(t)
	require.Len(t, project.Stages, 1)

	s2, err := project.NewStage("", "node1", "path")
	require.NoError(t, err)

	require.Len(t, project.Stages, 2)
	require.Equal(t, "node1", s2.Node().Name)
	require.Equal(t, project, s2.Project())
}

func TestProjectNewStageErrorPaths(t *testing.T) {
	project := testProject(t)
	require.Len(t, project.Stages, 1)

	s2, err := project.NewStage("my-stage", "node1", "path")
	require.Nil(t, s2)
	require.Error(t, err)

	var see *StageExistsError
	require.ErrorAs(t, err, &see)

	s2, err = project.NewStage("stage2", "not-found", "path")
	require.Nil(t, s2)
	require.Error(t, err)
	var nnfe *NodeNotFoundError
	require.ErrorAs(t, err, &nnfe)
}

func TestProjectRemoveStage(t *testing.T) {
	project := testProject(t)

	s2, err := project.NewStage("stage2", "node1", "path")
	require.NoError(t, err)
	require.NotEqual(t, s2, project.DefaultStage())

	require.Len(t, project.Stages, 2)
	project.RemoveStage("my-stage")
	require.Len(t, project.Stages, 1)
	require.Equal(t, s2, project.DefaultStage())
}

func TestProjectNumberOfNodes(t *testing.T) {
	project := testProject(t)

	require.Equal(t, project.AsCliProject().Nodes, 1)
	require.Equal(t, project.AsCliProject().AWSAccounts, 1)
	require.Equal(t, project.AsCliProject().AWSRegions, 1)

	project.NewStage("stage2", "node1", "path")
	require.Equal(t, 2, project.NumberOfStages())
	require.Equal(t, 1, project.AsCliProject().Nodes)
	require.Equal(t, project.AsCliProject().AWSAccounts, 1)
	require.Equal(t, project.AsCliProject().AWSRegions, 1)

	project.NewStage("stage3", "node2", "path")
	require.Equal(t, 3, project.NumberOfStages())
	require.Equal(t, 2, project.AsCliProject().Nodes)
	require.Equal(t, project.AsCliProject().AWSAccounts, 1)
	require.Equal(t, project.AsCliProject().AWSRegions, 1)

	project.NewStage("stage4", "node3", "path")
	require.Equal(t, 4, project.NumberOfStages())
	require.Equal(t, 3, project.AsCliProject().Nodes)
	require.Equal(t, project.AsCliProject().AWSAccounts, 2)
	require.Equal(t, project.AsCliProject().AWSRegions, 1)
}

func TestProjectSetDefaultStage(t *testing.T) {
	project := testProject(t)
	require.Equal(t, "my-stage", project.DefaultStage().Name)

	project.SetDefaultStage("non-existent")
	require.Equal(t, "my-stage", project.DefaultStage().Name)

	project.NewStage("stage2", "node2", "path")
	require.Equal(t, "my-stage", project.DefaultStage().Name)

	project.SetDefaultStage("stage2")
	require.Equal(t, "stage2", project.DefaultStage().Name)
	require.False(t, project.Stage("my-stage").Default)

	project.SetDefaultStage("my-stage")
	require.Equal(t, "my-stage", project.DefaultStage().Name)
	require.True(t, project.Stage("my-stage").Default)
	require.False(t, project.Stage("stage2").Default)
}

func TestProjectResourceNaming(t *testing.T) {
	project := testProject(t)

	require.Equal(t, "functions/my-project", project.FunctionsBucketPrefix())
	require.Equal(t, "state/my-project", project.StateBucketPrefix())
}

func testProject(t *testing.T) *Project {
	workspace := Workspace{
		ID: "my-workspace-id",
		Nodes: []*Node{
			{
				Name: "node1",
				ID:   "uid1",
			},
			{
				Name: "node2",
				ID:   "uid2",
			},
			{
				Name:      "node3",
				ID:        "uid2",
				AccountID: "2",
			},
		},
	}
	project := Project{
		Name: "my-project",
		Stages: []*Stage{
			{
				Name:     "my-stage",
				NodeName: "node1",
				Default:  true,
			},
		},
	}
	Factory(&workspace, &project, nil)
	return &project
}
