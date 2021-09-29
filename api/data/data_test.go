package data

import (
	"testing"

	"github.com/mantil-io/mantil/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataResponse(t *testing.T) {
	d := &Data{
		stage: &config.Stage{
			Name: "test-project",
		},
	}
	resp, err := d.data()
	require.NoError(t, err)
	assert.NotNil(t, resp.Stage)
}
