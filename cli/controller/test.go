package controller

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/shell"
)

type TestArgs struct {
	RunRegexp string
	Stage     string
}

func Test(a TestArgs) error {
	fs, stage, err := newStoreWithStage(a.Stage)
	if err != nil {
		return log.Wrap(err)
	}
	return runTests(fs.ProjectRoot(), stage.RestEndpoint(), a.RunRegexp)
}

func runTests(projectPath, apiURL, runRegexp string) error {
	args := []string{"go", "test", "-v"}
	if runRegexp != "" {
		args = append(args, "--run", runRegexp)
	}
	err := shell.Exec(shell.ExecOptions{
		Env:          []string{fmt.Sprintf("%s=%s", domain.EnvApiURL, apiURL)},
		Args:         args,
		WorkDir:      projectPath + "/test",
		Logger:       ui.Info,
		ShowShellCmd: false,
		ShowExitCode: false,
	})
	if err == nil {
		ui.Notice("PASS")
	} else {
		ui.ErrorLine("FAIL")
	}
	return nil
}
