package cmd

import (
	"regexp"
	"time"

	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"

	"github.com/mantil-io/mantil.go/pkg/shell"
	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/radovskyb/watcher"
)

type watchArgs struct {
	method string
	data   string
	test   bool
	stage  string
}

type watchCmd struct {
	deploy *deploy.Cmd
	invoke func() error
	path   string
	test   bool
}

func newWatch(a watchArgs) (*watchCmd, error) {
	fs, err := project.NewStoreWithStage(a.stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	stage := fs.Stage(a.stage)

	deploy, err := deploy.NewWithStage(fs, stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	var invoke func() error
	if a.method != "" {
		invoke = project.InvokeCallback(stage, a.method, a.data, false, true)
	}
	return &watchCmd{
		path:   fs.ProjectRoot(),
		deploy: deploy,
		invoke: invoke,
		test:   a.test,
	}, nil
}

func (c *watchCmd) run() error {
	c.watch(func() {
		ui.Info("")
		ui.Info("==> Changes detected")
		if err := c.deploy.Deploy(); err != nil {
			ui.Error(err)
			return
		}
		if !c.deploy.HasUpdates() {
			return
		}
		ui.Info("")
		if c.invoke != nil {
			ui.Info("==> Invoking function")
			if err := c.invoke(); err != nil {
				ui.Error(err)
			}
		}
		if c.test {
			ui.Info("")
			ui.Info("==> Running tests")
			err := shell.Exec(shell.ExecOptions{
				Args:    []string{"go", "test", "-v"},
				WorkDir: c.path + "/test",
				Logger:  ui.Info,
			})
			if err != nil {
				ui.Error(err)
			}
		}
	})
	return nil
}

func (c *watchCmd) watch(onChange func()) {
	w := watcher.New()

	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write, watcher.Create, watcher.Remove)

	// only watch for changes in go files
	r := regexp.MustCompile(`\.go$`)
	w.AddFilterHook(watcher.RegexFilterHook(r, false))

	go func() {
		for {
			select {
			case <-w.Event:
				onChange()
			case err := <-w.Error:
				ui.Fatal(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(c.path); err != nil {
		ui.Fatal(err)
	}

	ui.Info("Watching Go files in %s", c.path)
	if err := w.Start(1 * time.Second); err != nil {
		ui.Fatal(err)
	}
}
