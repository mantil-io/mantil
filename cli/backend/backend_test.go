package backend_test

import (
	"os"
	"strings"
	"testing"

	"github.com/mantil-io/mantil.go/pkg/streaming/nats"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/backend"
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
	var logs []string
	logSink := func(ch chan []byte) {
		for line := range ch {
			logs = append(logs, string(line))
		}
	}

	// test happy path
	err = backend.Lambda(aws.Lambda(), lambdaName, logSink).Call("logs", req, &rsp)
	require.NoError(t, err)
	require.Equal(t, rsp.Response, "Hello, Foo")
	//t.Logf("rsp %v", rsp)
	//t.Logf("logs: %#v", logs)
	require.Len(t, logs, 5)
	require.Equal(t, logs[0], "start")
	require.Equal(t, logs[1], "request name: Foo")
	require.Equal(t, logs[2], "request found")
	require.True(t, strings.HasPrefix(logs[3], "mantil-nats"))
	//require.True(t, strings.HasPrefix(logs[4], "mantil-nats"))
	require.Equal(t, logs[4], "end")

	// test server side error
	logs = make([]string, 0)
	req.Name = "Bar"
	err = backend.Lambda(aws.Lambda(), lambdaName, logSink).Call("logs", req, &rsp)
	require.Error(t, err)
	remoteErr := &nats.ErrRemoteError{}
	require.ErrorAs(t, err, &remoteErr)
	require.Equal(t, "name can't be Bar", remoteErr.Error())
	require.Len(t, logs, 3)
	require.Equal(t, logs[0], "start")
	require.Equal(t, logs[1], "request name: Bar")
	require.Equal(t, logs[2], "end")
	//t.Logf("error %s", remoteErr)
	//

	logs = make([]string, 0)
	err = backend.Lambda(aws.Lambda(), lambdaName, logSink).Call("ne-postoji", req, &rsp)
	require.Error(t, err)
	remoteErr = &nats.ErrRemoteError{}
	require.ErrorAs(t, err, &remoteErr)
	require.Equal(t, "method [ne-postoji] not found", remoteErr.Error())

	logs = make([]string, 0)
	err = backend.Lambda(aws.Lambda(), lambdaName+"a", logSink).Call("ne-postoji", req, &rsp)
	require.Error(t, err)
}
