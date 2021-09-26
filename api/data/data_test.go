package data

import (
	"testing"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataApi(t *testing.T) {
	// TODO: make sure resources neceessary for test are already created
	t.Skip()
	d := &Data{}
	req := &dto.DataRequest{
		ProjectName: "project-test", // TODO: project used for testing
	}
	err := d.init(req)
	require.NoError(t, err)
	assert.NotNil(t, d.project)

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
	assert.NotNil(t, resp.Project)
}
