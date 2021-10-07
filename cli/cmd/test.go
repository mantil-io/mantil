package cmd

import (
	"github.com/mantil-io/mantil.go/pkg/shell"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
)

type testFlags struct {
	runRegexp string
	stage     string
}

type testCmd struct {
	ctx       *project.Context
	runRegexp string
}

func newTest(f *testFlags) (*testCmd, error) {
	ctx, err := project.ContextWithStage(f.stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &testCmd{
		ctx:       ctx,
		runRegexp: f.runRegexp,
	}, nil
}

func (c *testCmd) run() error {
	stageURL, err := c.ctx.StageRestEndpoint()
	if err != nil {
		return err
	}
	return shell.Exec(shell.ExecOptions{
		Env:          []string{workspace.EnvApiURL + stageURL},
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
