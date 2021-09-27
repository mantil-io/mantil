package security

import (
	"testing"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type awsMock struct{}

func (a *awsMock) AccountID() (string, error) {
	return "123456789012", nil
}

func (a *awsMock) Region() string {
	return "region"
}

func (a *awsMock) RoleCredentials(name, role, policy string) (*aws.Credentials, error) {
	return &aws.Credentials{
		AccessKeyID:     "accessKeyID",
		SecretAccessKey: "secretAccessKey",
		SessionToken:    "sessionToken",
	}, nil
}

func TestSecurityApi(t *testing.T) {
	s := &Security{
		awsClient: &awsMock{},
		project: &config.Project{
			Name:   "test-project",
			Bucket: "test-bucket",
		},
		stage: &config.Stage{
			Name:        "test-stage",
			PublicSites: make([]*config.PublicSite, 0),
		},
	}
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
