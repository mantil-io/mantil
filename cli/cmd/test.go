package cmd

import (
	"github.com/mantil-io/mantil.go/pkg/shell"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
)

type testCmd struct {
	project   *config.Project
	stageName string
	repoPath  string
	runRegexp string
}

func (c *testCmd) run() error {
	err := shell.Exec(shell.ExecOptions{
		Env:          []string{"MANTIL_API_URL=" + c.project.RestEndpoint(c.stageName)},
		Args:         c.args(),
		WorkDir:      c.repoPath + "/test",
		Logger:       log.Info,
		ShowShellCmd: false,
	})
	if err != nil {
		return nil
	}
	return nil
}

func (c *testCmd) args() []string {
	args := []string{"go", "test", "-v"}
	if c.runRegexp != "" {
		args = append(args, "--run", c.runRegexp)
	}
	return args
}
