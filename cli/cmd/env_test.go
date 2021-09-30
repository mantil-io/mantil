package cmd

import (
	"testing"

	"github.com/mantil-io/mantil/workspace"
	"github.com/stretchr/testify/assert"
)

func TestEnvOutput(t *testing.T) {
	env := "env output"
	stage := &workspace.Stage{
		Endpoints: &workspace.StageEndpoints{
			Rest: "rest endpoint",
		},
	}
	cmd := &envCmd{}

	cmd.url = false
	assert.Equal(t, env, cmd.output(env, stage))

	cmd.url = true
	assert.Equal(t, stage.Endpoints.Rest, cmd.output(env, stage))
}
