package test

import (
	"strings"
	"testing"

	"github.com/mantil-io/mantil.go/logs"
	"github.com/mantil-io/mantil/cli/controller/invoke"
	"github.com/mantil-io/mantil/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPing(t *testing.T) {
	r := newCliRunnerWithWorkspaceCopy(t)
	t.Parallel()

	projectName := "my-ping"
	r.Assert("Your project is ready",
		"mantil", "new", projectName)
	r.logf("created %s project in %s", projectName, r.SetWorkdir(projectName))

	r.Assert("Deploy successful!",
		"mantil", "stage", "new", "test", "--node", defaultNodeName)

	r.Assert("No changes - nothing to deploy",
		"mantil", "deploy")

	r.Assert("pong",
		"mantil", "invoke", "ping")

	r.Assert("PASS",
		"mantil", "test")

	testAddLogsApi(r)
	r.WithWorkdir(func() { testBackendInvoke(t) })

	r.Assert("Stage test was successfully destroyed!",
		"mantil", "stage", "destroy", "test", "--yes")
}

func testAddLogsApi(r *cliRunner) {
	r.Assert(`Generating function logs`,
		"mantil", "generate", "api", "logs")

	r.CpToProject("./logs.go.txt", "api/logs/logs.go")

	r.Assert("Deploy successful!",
		"mantil", "deploy")

	c := r.Assert(`"Response": "Hello, Foo"`,
		"mantil", "invoke", "logs/test", "-d", `{"name": "Foo"}`)
	r.StdoutContains(c, "start")
	r.StdoutContains(c, "request name: Foo")
	r.StdoutContains(c, "request found")
	r.StdoutContains(c, "mantil-nats-config ->")
	r.StdoutContains(c, "end")

	c = r.Assert(`name can't be Bar`,
		"mantil", "invoke", "logs/test", "-d", `{"name": "Bar"}`)
	r.StdoutContains(c, "start")
	r.StdoutContains(c, "request name: Bar")
	r.StdoutContains(c, "end")
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
