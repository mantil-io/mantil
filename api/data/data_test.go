package data

import (
	"testing"

	"github.com/mantil-io/mantil/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataApi(t *testing.T) {
	d := &Data{
		stage: &config.Stage{
			Name: "test-project",
		},
	}
	tests := []func(*Data, *testing.T){
		testDataResponse,
	}

	for _, test := range tests {
		test(d, t)
	}
}

func testDataResponse(d *Data, t *testing.T) {
	resp, err := d.data()
	require.NoError(t, err)
	assert.NotNil(t, resp.Stage)
}
