package security

import (
	"testing"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityApi(t *testing.T) {
	// TODO: make sure resources neccessary for test are already created
	t.Skip()
	s := New()
	req := &dto.SecurityRequest{
		ProjectName: "project-test",  // TODO: project used for testing
		StageName:   "project-stage", // TODO: stage used for testing
	}
	err := s.init(req)
	require.NoError(t, err)
	assert.NotNil(t, s.awsClient)
	assert.NotNil(t, s.project)
	assert.NotNil(t, s.stage)

	tests := []func(*Security, *testing.T){
		testCliUserRole,
		testProjectCredentials,
		testSecurityResponse,
	}

	for _, test := range tests {
		test(s, t)
	}
}

func testCliUserRole(s *Security, t *testing.T) {
	role, err := s.cliUserRole()
	require.NoError(t, err)
	assert.NotEmpty(t, role)

}

func testProjectCredentials(s *Security, t *testing.T) {
	// create policy template data
	pptd, err := s.projectPolicyTemplateData()
	require.NoError(t, err)
	assert.NotEmpty(t, pptd.Name)
	assert.NotEmpty(t, pptd.Bucket)
	assert.NotEmpty(t, pptd.Region)
	assert.NotEmpty(t, pptd.AccountID)
	assert.NotNil(t, pptd.PublicSites)
	assert.NotEmpty(t, pptd.LogGroup)

	// render policy from template
	policy, err := s.executeProjectPolicyTemplate(pptd)
	require.NoError(t, err)
	assert.NotEmpty(t, policy)

	// generate credentials for policy
	creds, err := s.credentialsForPolicy(policy)
	require.NoError(t, err)
	assert.NotNil(t, creds)
	assert.NotEmpty(t, creds.AccessKeyID)
	assert.NotEmpty(t, creds.SecretAccessKey)
	assert.NotEmpty(t, creds.SessionToken)
	assert.NotEmpty(t, creds.Region)
}

func testSecurityResponse(s *Security, t *testing.T) {
	resp, err := s.credentials()
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessKeyID)
	assert.NotEmpty(t, resp.SecretAccessKey)
	assert.NotEmpty(t, resp.SessionToken)
	assert.NotEmpty(t, resp.Region)
}
