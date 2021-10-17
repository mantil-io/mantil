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
	ctx    *project.Context
	deploy *deploy.Cmd
	invoke *invokeCmd
	test   bool
	data   string
}

func newWatch(a watchArgs) (*watchCmd, error) {
	fs, err := project.NewStoreWithStage(a.stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	stage := fs.Stage(a.stage)

	// TODO remove ctx we already have fs, stage
	ctx, err := project.ContextWithStage(a.stage)
	if err != nil {
		return nil, log.Wrap(err)
	}

	deploy, err := deploy.NewWithStage(fs, stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	var invoke *invokeCmd
	if a.method != "" {
		invoke = &invokeCmd{
			ctx:            ctx,
			path:           a.method,
			data:           a.data,
			includeHeaders: false,
			includeLogs:    true,
		}
	}
	return &watchCmd{
		ctx:    ctx,
		deploy: deploy,
		invoke: invoke,
		test:   a.test,
		data:   a.data,
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
			if err := c.invoke.run(); err != nil {
				ui.Error(err)
			}
		}
		if c.test {
			ui.Info("")
			ui.Info("==> Running tests")
			err := shell.Exec(shell.ExecOptions{
				Args:    []string{"go", "test", "-v"},
				WorkDir: c.ctx.Path + "/test",
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

	if err := w.AddRecursive(c.ctx.Path); err != nil {
		ui.Fatal(err)
	}

	ui.Info("Watching Go files in %s", c.ctx.Path)
	if err := w.Start(1 * time.Second); err != nil {
		ui.Fatal(err)
	}
}
