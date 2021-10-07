package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/shell"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	cli := aws.NewForTests(t)
	if cli == nil {
		t.Skip("skip: AWS client not initialized")
	}

	tmpDir, err := ioutil.TempDir("/tmp", "mantil-tests-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	t.Logf("using tmp dir: %s", tmpDir)

	accountID, err := cli.AccountID()
	require.NoError(t, err)
	t.Logf("using accountID: %s", accountID)

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
	run("destroy stage", pingDir, "mantil", "stage", "destroy", "test", "--force")

	run("uninstall", tmpDir, "mantil", "aws", "uninstall", "--aws-profile", profile)
}

func showShellOut(t *testing.T, args ...string) {
	out, err := shell.Output(shell.ExecOptions{Args: args})
	require.NoError(t, err)
	cmd := strings.Join(args, " ")
	t.Logf("\n$ %s\n%s", cmd, out)
}
