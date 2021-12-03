package test

import (
	"strings"
	"testing"

	"github.com/mantil-io/mantil.go/logs"
	"github.com/mantil-io/mantil/cli/controller/invoke"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/clitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {
	r := newCliRunnerWithWorkspaceCopy(t)

	c := clitest.New(t).
		Env(domain.EnvWorkspacePath, r.TestDir()).
		Workdir(r.TestDir())
	t.Parallel()

	projectName := "my-ping"
	c.Run("mantil", "new", projectName).
		Contains("Your project is ready")

	t.Logf("created %s project in %s", projectName, c.Cd(projectName))

	c.Run("mantil", "stage", "new", "test", "--node", defaultNodeName).
		Contains("Deploy successful!")
	c.Run("mantil", "deploy").Contains("No changes - nothing to deploy")
	c.Run("mantil", "invoke", "ping").Contains("pong")
	c.Run("mantil", "test").Contains("PASS")

	testAddLogsApi(c)
	c.WithWorkdir(func() { testBackendInvoke(t) })

	c.Run("mantil", "stage", "destroy", "test", "--yes").
		Contains("Stage test was successfully destroyed!")
}

func testAddLogsApi(r *clitest.Env) {
	r.Run("mantil", "generate", "api", "logs").
		Contains(`Generating function logs`)

	r.CpToWorkdir("./logs.go.txt", "api/logs/logs.go")

	r.Run("mantil", "deploy").
		Contains("Deploy successful!")

	r.Run("mantil", "invoke", "logs/test", "-d", `{"name": "Foo"}`).
		Contains(`"Response": "Hello, Foo"`).
		Contains("start").
		Contains("request name: Foo").
		Contains("request found").
		Contains("mantil-nats-config ->").
		Contains("end")

	r.Run("mantil", "invoke", "logs/test", "-d", `{"name": "Bar"}`).
		Contains(`name can't be Bar`).
		Contains("start").
		Contains("request name: Bar").
		Contains("end")
}

func testBackendInvoke(t *testing.T) {
	lambdaName := getLambdaFunctionName(t, "logs")
	require.NotEmpty(t, lambdaName)

	// request and response
	req := struct {
		Name string
	}{Name: "Foo"}
	var rsp struct {
		Response string
	}

	// collect log lines
	var logLines []string
	logSink := func(ch chan []byte) {
		for line := range ch {
			logLines = append(logLines, string(line))
		}
	}

	// test happy path
	err := invoke.Lambda(awsCli.Lambda(), lambdaName, logSink).Do("test", req, &rsp)
	assert.NoError(t, err)
	assert.Equal(t, rsp.Response, "Hello, Foo")
	assert.Len(t, logLines, 5)
	assert.True(t, strings.HasSuffix(logLines[0], "start"))
	assert.True(t, strings.HasSuffix(logLines[1], "request name: Foo"))
	assert.True(t, strings.HasSuffix(logLines[2], "request found"))
	assert.True(t, strings.Contains(logLines[3], "mantil-nats"))
	assert.True(t, strings.HasSuffix(logLines[4], "end"))

	// test server side error
	logLines = make([]string, 0)
	req.Name = "Bar"
	err = invoke.Lambda(awsCli.Lambda(), lambdaName, logSink).Do("test", req, &rsp)
	assert.Error(t, err)
	remoteErr := &logs.ErrRemoteError{}
	assert.ErrorAs(t, err, &remoteErr)
	assert.Equal(t, "name can't be Bar", remoteErr.Error())
	assert.Len(t, logLines, 3)
	assert.True(t, strings.HasSuffix(logLines[0], "start"))
	assert.True(t, strings.HasSuffix(logLines[1], "request name: Bar"))
	assert.True(t, strings.HasSuffix(logLines[2], "end"))

	// try the method which don't exists
	logLines = make([]string, 0)
	err = invoke.Lambda(awsCli.Lambda(), lambdaName, logSink).Do("ne-postoji", req, &rsp)
	assert.Error(t, err)
	remoteErr = &logs.ErrRemoteError{}
	assert.ErrorAs(t, err, &remoteErr)
	assert.Equal(t, "method [ne-postoji] not found", remoteErr.Error())

	// try the lambda function which don't exists
	logLines = make([]string, 0)
	err = invoke.Lambda(awsCli.Lambda(), lambdaName+"a", logSink).Do("ne-postoji", req, &rsp)
	assert.Error(t, err)
}

func getLambdaFunctionName(t *testing.T, strContains string) string {
	fs, err := domain.NewSingleDeveloperProjectStore()
	require.NoError(t, err)
	p := fs.Project()
	require.NotNil(t, p)
	s := p.DefaultStage()
	require.NotNil(t, s)
	for _, f := range s.Functions {
		if strings.Contains(f.LambdaName(), strContains) {
			return f.LambdaName()
		}
	}
	return ""
}
