package cmd

import (
	"testing"

	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvOutput(t *testing.T) {
	cmd := &envCmd{
		ctx: &project.Context{},
	}

	_, err := cmd.output()
	require.Error(t, err)

	cmd.ctx = &project.Context{
		Project: &workspace.Project{
			Name: "project",
		},
		Stage: &workspace.Stage{
			Name: "stage",
			Endpoints: &workspace.StageEndpoints{
				Rest: "stageRestURL",
			},
		},
	}

	output, err := cmd.output()
	require.NoError(t, err)
	stageURL, err := cmd.ctx.StageRestEndpoint()
	require.NoError(t, err)
	assert.Equal(t, workspace.Env(cmd.ctx.Project.Name, stageURL), output)

	cmd.url = true
	output, err = cmd.output()
	require.NoError(t, err)
	stageURL, err = cmd.ctx.StageRestEndpoint()
	require.NoError(t, err)
	require.Equal(t, stageURL, output)
}
