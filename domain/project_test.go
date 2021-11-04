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

	_, err := project.NewStage("stage2", "node1")
	require.NoError(t, err)

	ds = project.DefaultStage()
	require.NotNil(t, ds)
	require.Equal(t, "my-stage", ds.Name)
}

func TestProjectNewStage(t *testing.T) {
	project := testProject(t)
	require.Len(t, project.Stages, 1)

	s2, err := project.NewStage("stage2", "node1")
	require.NoError(t, err)

	require.Len(t, project.Stages, 2)
	require.Equal(t, "node1", s2.Node().Name)
	require.Equal(t, project, s2.Project())
	require.False(t, s2.Default)
}

func TestProjectFirstNewStageIsDefault(t *testing.T) {
	project := testProject(t)
	require.Len(t, project.Stages, 1)
	project.RemoveStage("my-stage")
	require.Len(t, project.Stages, 0)

	s2, err := project.NewStage("stage2", "node1")
	require.NoError(t, err)

	require.Equal(t, "node1", s2.Node().Name)
	require.Equal(t, project, s2.Project())
	require.True(t, s2.Default)
}

func TestProjectNewStageWithEmptyName(t *testing.T) {
	project := testProject(t)
	require.Len(t, project.Stages, 1)

	s2, err := project.NewStage("", "node1")
	require.NoError(t, err)

	require.Len(t, project.Stages, 2)
	require.Equal(t, "node1", s2.Node().Name)
	require.Equal(t, project, s2.Project())
}

func TestProjectNewStageErrorPaths(t *testing.T) {
	project := testProject(t)
	require.Len(t, project.Stages, 1)

	s2, err := project.NewStage("my-stage", "node1")
	require.Nil(t, s2)
	require.Error(t, err)

	var see *StageExistsError
	require.ErrorAs(t, err, &see)

	s2, err = project.NewStage("stage2", "not-found")
	require.Nil(t, s2)
	require.Error(t, err)
	var nnfe *NodeNotFoundError
	require.ErrorAs(t, err, &nnfe)
}

func TestProjectRemoveStage(t *testing.T) {
	project := testProject(t)

	s2, err := project.NewStage("stage2", "node1")
	require.NoError(t, err)
	require.NotEqual(t, s2, project.DefaultStage())

	require.Len(t, project.Stages, 2)
	project.RemoveStage("my-stage")
	require.Len(t, project.Stages, 1)
	require.Equal(t, s2, project.DefaultStage())
}

func testProject(t *testing.T) *Project {
	workspace := Workspace{
		Name: "my-workspace",
		Nodes: []*Node{
			{
				Name: "node1",
				UID:  "suffix",
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
