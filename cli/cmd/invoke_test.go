package cmd

import (
	"testing"

	"github.com/mantil-io/mantil/cli/commands"
	"github.com/mantil-io/mantil/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvokeURL(t *testing.T) {
	cmd := &invokeCmd{
		endpoint: "endpoint",
		ctx: &commands.ProjectContext{
			Stage: &config.Stage{
				Name: "stage",
				Endpoints: &config.StageEndpoints{
					Rest: "stageRestURL",
				},
			},
		},
	}

	cmd.ctx.Project = &config.Project{}
	_, err := cmd.url()
	require.Error(t, err)

	cmd.ctx.Project.Stages = []*config.Stage{cmd.ctx.Stage}
	url, err := cmd.url()
	require.NoError(t, err)
	assert.Equal(t, "stageRestURL/endpoint", url)
}
