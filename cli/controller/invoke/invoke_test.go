package invoke_test

import (
	"os"
	"strings"
	"testing"

	"github.com/mantil-io/mantil.go/logs"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/controller/invoke"
	"github.com/mantil-io/mantil/domain"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	t.Skip("used in development, moved to end_to_end")
	err := os.Chdir("/Users/ianic/mantil-io/ping")
	require.NoError(t, err)

	// get lambda function name
	fs, err := domain.NewSingleDeveloperProjectStore()
	require.NoError(t, err)
	p := fs.Project()
	require.NotNil(t, p)
	s := p.DefaultStage()
	require.NotNil(t, s)
	lambdaName := s.Functions[0].LambdaName()
	t.Logf("lambda function name %s", lambdaName)

	// connect to aws
	aws, err := aws.NewFromProfile("org5")
	require.NoError(t, err)

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
	err = invoke.Lambda(aws.Lambda(), lambdaName, logSink).Do("logs", req, &rsp)
	require.NoError(t, err)
	require.Equal(t, rsp.Response, "Hello, Foo")
	//t.Logf("rsp %v", rsp)
	//t.Logf("logs: %#v", logs)
	require.Len(t, logLines, 5)
	require.Equal(t, logLines[0], "start")
	require.Equal(t, logLines[1], "request name: Foo")
	require.Equal(t, logLines[2], "request found")
	require.True(t, strings.HasPrefix(logLines[3], "mantil-nats"))
	//require.True(t, strings.HasPrefix(logs[4], "mantil-nats"))
	require.Equal(t, logLines[4], "end")

	// test server side error
	logLines = make([]string, 0)
	req.Name = "Bar"
	err = invoke.Lambda(aws.Lambda(), lambdaName, logSink).Do("logs", req, &rsp)
	require.Error(t, err)
	remoteErr := &logs.ErrRemoteError{}
	require.ErrorAs(t, err, &remoteErr)
	require.Equal(t, "name can't be Bar", remoteErr.Error())
	require.Len(t, logLines, 3)
	require.Equal(t, logLines[0], "start")
	require.Equal(t, logLines[1], "request name: Bar")
	require.Equal(t, logLines[2], "end")
	//t.Logf("error %s", remoteErr)
	//

	logLines = make([]string, 0)
	err = invoke.Lambda(aws.Lambda(), lambdaName, logSink).Do("ne-postoji", req, &rsp)
	require.Error(t, err)
	remoteErr = &logs.ErrRemoteError{}
	require.ErrorAs(t, err, &remoteErr)
	require.Equal(t, "method [ne-postoji] not found", remoteErr.Error())

	logLines = make([]string, 0)
	err = invoke.Lambda(aws.Lambda(), lambdaName+"a", logSink).Do("ne-postoji", req, &rsp)
	require.Error(t, err)
}
