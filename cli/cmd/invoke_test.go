package cmd

import (
	"testing"

	"github.com/mantil-io/mantil/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvokeURL(t *testing.T) {
	cmd := &invokeCmd{
		endpoint:  "endpoint",
		stageName: "stage",
	}

	cmd.project = &config.Project{}
	_, err := cmd.url()
	require.Error(t, err)

	cmd.project.Stages = []*config.Stage{
		{
			Name: "stage",
			Endpoints: &config.StageEndpoints{
				Rest: "stageRestURL",
			},
		},
	}
	url, err := cmd.url()
	require.NoError(t, err)
	assert.Equal(t, "stageRestURL/endpoint", url)
}
