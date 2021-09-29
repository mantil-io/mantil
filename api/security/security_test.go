package security

import (
	"io/ioutil"
	"testing"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/config"
	"github.com/sergi/go-diff/diffmatchpatch"
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

func TestCliUserRole(t *testing.T) {
	s := &Security{
		awsClient: &awsMock{},
	}
	role, err := s.cliUserRole()
	require.NoError(t, err)
	assert.NotEmpty(t, role)

}

func TestProjectCredentialsWithoutStage(t *testing.T) {
	s := &Security{
		req: &SecurityRequest{
			ProjectName: "test-project",
			StageName:   "test-stage",
		},
		bucketName: "bucket",
		awsClient:  &awsMock{},
	}
	pptd, err := s.projectPolicyTemplateData()
	require.NoError(t, err)
	assert.NotEmpty(t, pptd.Name)
	assert.NotEmpty(t, pptd.Bucket)
	assert.NotEmpty(t, pptd.Region)
	assert.NotEmpty(t, pptd.AccountID)
	assert.Nil(t, pptd.PublicSites)
	assert.Empty(t, pptd.LogGroup)

	policy, err := s.executeProjectPolicyTemplate(pptd)
	require.NoError(t, err)

	policyWithoutStage, err := ioutil.ReadFile("testdata/policy-no-stage")
	require.NoError(t, err)
	compareStrings(t, string(policyWithoutStage), policy)

	creds, err := s.credentialsForPolicy(policy)
	require.NoError(t, err)
	assert.NotEmpty(t, creds.AccessKeyID)
	assert.NotEmpty(t, creds.SecretAccessKey)
	assert.NotEmpty(t, creds.SessionToken)
	assert.NotEmpty(t, creds.Region)
}

func TestProjectCredentialsWithStage(t *testing.T) {
	s := &Security{
		req: &SecurityRequest{
			ProjectName: "test-project",
			StageName:   "test-stage",
		},
		stage: &config.Stage{
			Name: "test-stage",
			PublicSites: []*config.PublicSite{
				{Bucket: "publicSite1"},
				{Bucket: "publicSite2"},
			},
		},
		bucketName: "bucket",
		awsClient:  &awsMock{},
	}
	pptd, err := s.projectPolicyTemplateData()
	require.NoError(t, err)
	assert.NotEmpty(t, pptd.Name)
	assert.NotEmpty(t, pptd.Bucket)
	assert.NotEmpty(t, pptd.Region)
	assert.NotEmpty(t, pptd.AccountID)
	assert.NotNil(t, pptd.PublicSites)
	assert.NotEmpty(t, pptd.LogGroup)

	policy, err := s.executeProjectPolicyTemplate(pptd)
	require.NoError(t, err)

	policyWithStage, err := ioutil.ReadFile("testdata/policy-stage")
	require.NoError(t, err)
	compareStrings(t, string(policyWithStage), policy)

	creds, err := s.credentialsForPolicy(policy)
	require.NoError(t, err)
	assert.NotEmpty(t, creds.AccessKeyID)
	assert.NotEmpty(t, creds.SecretAccessKey)
	assert.NotEmpty(t, creds.SessionToken)
	assert.NotEmpty(t, creds.Region)

}

func compareStrings(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Logf("diff of strings")
		t.Logf("expected \n%s, actual \n%s", expected, actual)
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expected, actual, false)
		t.Logf("diff: \n%s", dmp.DiffPrettyText(diffs))
		t.Fatalf("failed")
	}
}
