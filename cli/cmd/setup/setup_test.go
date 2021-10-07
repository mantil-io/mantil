package setup

import (
	"io/ioutil"
	"testing"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/build"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTemplate(t *testing.T) {
	td := stackTemplateData{
		Name:   "mantil-setup",
		Bucket: "bucket",
		S3Key:  "bucket-key",
		Region: "region",
	}
	actual, err := renderStackTemplate(td)
	require.NoError(t, err)
	expected, err := ioutil.ReadFile("testdata/template.yml")
	require.NoError(t, err)
	equalStrings(t, string(expected), actual)
}

func TestCreateLambda(t *testing.T) {
	cli := aws.NewForTests(t)
	if cli == nil {
		t.Skip("skip: AWS client not initialized")
	}
	cmd := new(cli)
	// empty at start
	alreadyRun, err := cmd.backendExists()
	require.NoError(t, err)
	require.False(t, alreadyRun)
	// create lambda
	err = cmd.ensureLambdaExists()
	require.NoError(t, err)
	// check exists
	exists, err := cmd.awsClient.LambdaExists(lambdaName)
	require.NoError(t, err)
	require.True(t, exists)
	// and one more
	alreadyRun, err = cmd.backendExists()
	require.NoError(t, err)
	require.True(t, alreadyRun)
	// clanup
	err = cmd.deleteLambda()
	require.NoError(t, err)
	// check we are at clean
	alreadyRun, err = cmd.backendExists()
	require.NoError(t, err)
	require.False(t, alreadyRun)
}

func TestCreateAndInvoke(t *testing.T) {
	cli := aws.NewForTests(t)
	if cli == nil {
		t.Skip("skip: AWS client not initialized")
	}
	cmd := new(cli)

	// empty at start
	alreadyRun, err := cmd.backendExists()
	require.NoError(t, err)
	require.False(t, alreadyRun)
	// create backend lambda functions
	ac, err := cmd.create()
	require.NoError(t, err)
	require.NotNil(t, ac)
	// check config
	assert.Equal(t, "dev", ac.Name)
	assert.NotNil(t, ac.Keys)
	assert.NotNil(t, ac.Endpoints)
	assert.NotEmpty(t, ac.Keys.Public)
	assert.NotEmpty(t, ac.Keys.Private)
	assert.NotEmpty(t, ac.Endpoints.Rest)
	assert.NotEmpty(t, ac.Endpoints.Ws)
	// show config
	buf, _ := ac.Marshal()
	t.Logf("account config: %s\n", buf)
	// cleanup
	err = cmd.destroy()
	require.NoError(t, err)
}

func equalStrings(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Logf("diff of strings")
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expected, actual, false)
		t.Logf("diff: \n%s", dmp.DiffPrettyText(diffs))
		t.Fatalf("failed")
	}
}

func new(cli *aws.AWS) *Cmd {
	v := build.Version()
	return &Cmd{
		awsClient:       cli,
		functionsBucket: v.FunctionsBucket(cli.Region()),
		functionsPath:   v.FunctionsPath(),
		accountName:     "dev",
	}
}
