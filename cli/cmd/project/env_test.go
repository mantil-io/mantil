package project

import (
	"testing"

	"github.com/mantil-io/mantil/workspace"
	"github.com/stretchr/testify/require"
)

func TestEnvOutput(t *testing.T) {
	var f EnvFlags

	out, err := Env(f)
	require.Error(t, err)
	require.Equal(t, "", out)

	ctx := &Context{
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

	out, err = ctx.env(EnvFlags{Stage: "stage"})
	require.NoError(t, err)
	require.Equal(t, out, `export MANTIL_PROJECT_NAME='project'
export MANTIL_API_URL='stageRestURL'
`)

	out, err = ctx.env(EnvFlags{Stage: "stage", Url: true})
	require.NoError(t, err)
	require.Equal(t, "stageRestURL", out)
}
