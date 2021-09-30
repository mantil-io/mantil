package cmd

import (
	"github.com/mantil-io/mantil/cli/cmd/project"
	"testing"

	"github.com/mantil-io/mantil/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvokeURL(t *testing.T) {
	cmd := &invokeCmd{
		endpoint: "endpoint",
		ctx: &project.Context{
			Stage: &workspace.Stage{
				Name: "stage",
				Endpoints: &workspace.StageEndpoints{
					Rest: "stageRestURL",
				},
			},
		},
	}

	cmd.ctx.Project = &workspace.Project{}
	_, err := cmd.url()
	require.Error(t, err)

	cmd.ctx.Project.Stages = []*workspace.Stage{cmd.ctx.Stage}
	url, err := cmd.url()
	require.NoError(t, err)
	assert.Equal(t, "stageRestURL/endpoint", url)
}
