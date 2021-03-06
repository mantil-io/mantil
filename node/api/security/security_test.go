package security

import (
	"flag"
	"io/fs"
	"io/ioutil"
	"testing"
	"time"

	"github.com/mantil-io/mantil/kit/aws"
	"github.com/mantil-io/mantil/node/dto"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update expected files")

type awsMock struct{}

func (a *awsMock) AccountID() string {
	return "123456789012"
}

func (a *awsMock) Region() string {
	return "region"
}

func (a *awsMock) RoleCredentials(name, role, policy string, durationSeconds int32) (*aws.Credentials, error) {
	return &aws.Credentials{
		AccessKeyID:     "accessKeyID",
		SecretAccessKey: "secretAccessKey",
		SessionToken:    "sessionToken",
		Expiration:      &time.Time{},
	}, nil
}

func TestProjectPolicyWithLogGroup(t *testing.T) {
	s := &Security{
		SecurityRequest: dto.SecurityRequest{
			CliRole:         "cliRole",
			Buckets:         []string{"bucket1", "bucket2", ""},
			LogGroupsPrefix: "logGroupsPrefix",
		},
		awsClient: &awsMock{},
	}
	pptd := s.projectPolicyTemplateData()
	assert.NotEmpty(t, pptd.Buckets)
	assert.NotEmpty(t, pptd.LogGroupsPrefix)
	assert.NotEmpty(t, pptd.Region)
	assert.NotEmpty(t, pptd.AccountID)

	policy, err := s.executeProjectPolicyTemplate(pptd)
	require.NoError(t, err)

	compare(t, "testdata/policy-log-group", policy)
}

func TestProjectPolicyWithoutLogGroup(t *testing.T) {
	s := &Security{
		SecurityRequest: dto.SecurityRequest{
			CliRole: "cliRole",
			Buckets: []string{"bucket1", "bucket2", ""},
		},
		awsClient: &awsMock{},
	}
	pptd := s.projectPolicyTemplateData()
	assert.NotEmpty(t, pptd.Buckets)
	assert.Empty(t, pptd.LogGroupsPrefix)
	assert.NotEmpty(t, pptd.Region)
	assert.NotEmpty(t, pptd.AccountID)

	policy, err := s.executeProjectPolicyTemplate(pptd)
	require.NoError(t, err)

	compare(t, "testdata/policy", policy)
}

func compare(t *testing.T, expectedFilename, policy string) {
	if *update {
		err := ioutil.WriteFile(expectedFilename, []byte(policy), fs.ModePerm)
		require.NoError(t, err)
		t.Logf("updated expected file %s", expectedFilename)
		return
	}
	expected, err := ioutil.ReadFile(expectedFilename)
	require.NoError(t, err)
	compareStrings(t, string(expected), policy)
}

func compareStrings(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Logf("diff of strings")
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expected, actual, false)
		t.Logf("diff: \n%s", dmp.DiffPrettyText(diffs))
		t.Fatalf("failed")
	}
}
