package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mantil-io/mantil.go/pkg/streaming/nats"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/backend"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/shell"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	cli := aws.NewForTests(t)
	if cli == nil {
		t.Skip("skip: AWS client not initialized")
	}

	workspacePath, err := ioutil.TempDir("/tmp", "mantil-workspace-")
	require.NoError(t, err)
	t.Setenv(domain.EnvWorkspacePath, workspacePath)
	t.Logf("setting workspace path to %s", workspacePath)

	tmpDir, err := ioutil.TempDir("/tmp", "mantil-tests-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	t.Logf("using tmp dir: %s", tmpDir)
	t.Logf("using accountID: %s", cli.AccountID())

	profile := aws.TestProfile()
	t.Logf("using AWS profile: %s", profile)

	// run shell command as Go sub test
	run := func(name, workDir string, args ...string) {
		t.Run(name, func(t *testing.T) {
			err := shell.Exec(shell.ExecOptions{
				Args:         args,
				WorkDir:      workDir,
				ShowShellCmd: true,
				Logger: func(format string, v ...interface{}) {
					t.Logf(format, v...)
				},
			})
			require.NoError(t, err)
		})
	}

	run("deploy", "./scripts", "./deploy.sh")
	t.Run("show environment", func(t *testing.T) {
		showShellOut(t, "which", "mantil")
		showShellOut(t, "mantil", "--version")
	})

	run("install", tmpDir, "mantil", "aws", "install", "--aws-profile", profile)

	pingDir := tmpDir + "/my-ping"
	run("create ping project", tmpDir, "mantil", "new", "my-ping")
	run("create stage", pingDir, "mantil", "stage", "new", "test")
	run("deploy ping project", pingDir, "mantil", "deploy")
	run("invoke method", pingDir, "mantil", "invoke", "ping")
	run("test ping project", pingDir, "mantil", "test")

	// invoke method tests
	run("generate logs api", pingDir, "mantil", "generate", "api", "logs")
	t.Run("update logs api", func(t *testing.T) { addLogsApi(t, pingDir) })
	run("show", pingDir, "cat", "api/logs/logs.go")
	run("deploy ping project", pingDir, "mantil", "deploy")
	run("invoke method", pingDir, "mantil", "invoke", "logs/test", "-l", "-d", `{"name": "Foo"}`)
	t.Run("backend invoke lambda function", func(t *testing.T) { testBackendInvoke(t, pingDir) })

	run("destroy stage", pingDir, "mantil", "stage", "destroy", "test", "--force")
	run("uninstall", tmpDir, "mantil", "aws", "uninstall", "--aws-profile", profile)
}

func showShellOut(t *testing.T, args ...string) {
	out, err := shell.Output(shell.ExecOptions{Args: args})
	require.NoError(t, err)
	cmd := strings.Join(args, " ")
	t.Logf("\n$ %s\n%s", cmd, out)
}

func addLogsApi(t *testing.T, pingDir string) {
	// used during development of both repos
	// err := appendToFile(pingDir, "go.mod", "\nreplace github.com/mantil-io/mantil.go => /Users/ianic/mantil-io/mantil.go\n")
	// require.NoError(t, err)

	err := replaceFile(pingDir, "api/logs/logs.go", `
package logs

import (
	"context"
	"fmt"
    "log"
    "github.com/mantil-io/mantil.go"
)


type Logs struct{}

func New() *Logs {
	return &Logs{}
}

func (l *Logs) Default(ctx context.Context) error {
	panic("not implemented")
}


type Request struct {
	Name string
}

type Response struct {
	Response string
}

func (l *Logs) Test(ctx context.Context, req *Request) (*Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request not found")
	}
	log.Printf("start")
	defer log.Printf("end")
	log.Printf("request name: %s", req.Name)
	if req.Name == "Bar" {
		return nil, fmt.Errorf("name can't be Bar")
	}
	mc, ok := mantil.FromContext(ctx)
	if ok {
		log.Printf("request found")
		for k, v := range mc.Request.Headers {
			log.Printf("%s -> %s", k, v)
		}
	}
	return &Response{Response: "Hello, " + req.Name}, nil
}
`)
	require.NoError(t, err)
}

func appendToFile(dir, file, content string) error {
	fn := filepath.Join(dir, file)
	f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		return err
	}
	return f.Close()
}

func replaceFile(dir, file, content string) error {
	fn := filepath.Join(dir, file)
	f, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(content); err != nil {
		return err
	}
	return f.Close()
}

func testBackendInvoke(t *testing.T, pingDir string) {
	err := os.Chdir(pingDir)
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
	if !strings.Contains(lambdaName, "logs") {
		lambdaName = s.Functions[1].LambdaName()
		t.Logf("lambda function name 1 %s", lambdaName)
	}

	// connect to aws
	aws := aws.NewForTests(t)
	if aws == nil {
		return
	}

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
	err = backend.Lambda(aws.Lambda(), lambdaName, logSink).Call("test", req, &rsp)
	require.NoError(t, err)
	require.Equal(t, rsp.Response, "Hello, Foo")
	//t.Logf("rsp %v", rsp)
	//t.Logf("logs: %#v", logs)
	require.Len(t, logs, 5)
	require.Equal(t, logs[0], "start")
	require.Equal(t, logs[1], "request name: Foo")
	require.Equal(t, logs[2], "request found")
	require.True(t, strings.HasPrefix(logs[3], "mantil-nats"))
	require.Equal(t, logs[4], "end")

	// test server side error
	logs = make([]string, 0)
	req.Name = "Bar"
	err = backend.Lambda(aws.Lambda(), lambdaName, logSink).Call("test", req, &rsp)
	require.Error(t, err)
	remoteErr := &nats.ErrRemoteError{}
	require.ErrorAs(t, err, &remoteErr)
	require.Equal(t, "name can't be Bar", remoteErr.Error())
	require.Len(t, logs, 3)
	require.Equal(t, logs[0], "start")
	require.Equal(t, logs[1], "request name: Bar")
	require.Equal(t, logs[2], "end")

	// try the method which don't exists
	logs = make([]string, 0)
	err = backend.Lambda(aws.Lambda(), lambdaName, logSink).Call("ne-postoji", req, &rsp)
	require.Error(t, err)
	remoteErr = &nats.ErrRemoteError{}
	require.ErrorAs(t, err, &remoteErr)
	require.Equal(t, "method [ne-postoji] not found", remoteErr.Error())

	// try the lambda function which don't exists
	logs = make([]string, 0)
	err = backend.Lambda(aws.Lambda(), lambdaName+"a", logSink).Call("ne-postoji", req, &rsp)
	require.Error(t, err)
}
