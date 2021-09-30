package cmd

import (
	"github.com/mantil-io/mantil.go/pkg/shell"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/workspace"
)

type testCmd struct {
	project   *workspace.Project
	stageName string
	repoPath  string
	runRegexp string
}

func (c *testCmd) run() error {
	return shell.Exec(shell.ExecOptions{
		//TODO: move magic string to constant
		Env:          []string{"MANTIL_API_URL=" + c.project.RestEndpoint(c.stageName)},
		Args:         c.args(),
		WorkDir:      c.repoPath + "/test",
		Logger:       log.Info,
		ShowShellCmd: false,
	})
}

func (c *testCmd) args() []string {
	args := []string{"go", "test", "-v"}
	if c.runRegexp != "" {
		args = append(args, "--run", c.runRegexp)
	}
	return args
}
