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
	stageURL  string
	path      string
	runRegexp string
}

func newTest(a testArgs) (*testCmd, error) {
	fs, err := project.NewStoreWithStage(a.stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &testCmd{
		stageURL:  fs.Stage(a.stage).Endpoints.Rest,
		path:      fs.ProjectRoot(),
		runRegexp: a.runRegexp,
	}, nil
}

func (c *testCmd) run() error {
	return shell.Exec(shell.ExecOptions{
		Env:          []string{fmt.Sprintf("%s=%s", workspace.EnvApiURL, c.stageURL)},
		Args:         c.args(),
		WorkDir:      c.path + "/test",
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
