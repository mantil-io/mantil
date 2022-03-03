package controller

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGithubIntegrationWorkflow(t *testing.T) {
	actual, err := renderIntegrationWorkflowTemplate(integrationWorkflowTemplate, "integration")
	require.NoError(t, err)
	expected, err := ioutil.ReadFile("testdata/integration_workflow.yml")
	require.NoError(t, err)
	equalStrings(t, string(expected), string(actual))
}
