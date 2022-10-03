package controller

import (
	"flag"
	"testing"

	"github.com/mantil-io/mantil/kit/testutil"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update expected files")

func TestRenderTemplate(t *testing.T) {
	td := stackTemplateData{
		Name:               "mantil-setup",
		Bucket:             "bucket",
		S3Key:              "bucket-key",
		Region:             "region",
		Suffix:             "suffix",
		APIGatewayLogsRole: "logs-role",
		Env: map[string]string{
			"key": "value",
		},
	}
	s := &Setup{}
	actual, err := s.renderStackTemplate(td)
	require.NoError(t, err)
	testutil.EqualFiles(t, "testdata/template.yml", string(actual), *update)
}
