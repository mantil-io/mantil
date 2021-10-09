package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil.go/pkg/shell"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
)

type testArgs struct {
	runRegexp string
	stage     string
}

type testCmd struct {
	ctx       *project.Context
	runRegexp string
}

func newTest(a testArgs) (*testCmd, error) {
	ctx, err := project.ContextWithStage(a.stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &testCmd{
		ctx:       ctx,
		runRegexp: a.runRegexp,
	}, nil
}

func (c *testCmd) run() error {
	stageURL, err := c.ctx.StageRestEndpoint()
	if err != nil {
		return err
	}
	return shell.Exec(shell.ExecOptions{
		Env:          []string{fmt.Sprintf("%s=%s", workspace.EnvApiURL, stageURL)},
		Args:         c.args(),
		WorkDir:      c.ctx.Path + "/test",
		Logger:       ui.Info,
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
