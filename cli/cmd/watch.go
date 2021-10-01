package cmd

import (
	"regexp"
	"time"

	"github.com/mantil-io/mantil/cli/cmd/project"

	"github.com/mantil-io/mantil.go/pkg/shell"
	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/radovskyb/watcher"
)

type watchCmd struct {
	ctx    *project.Context
	deploy *deploy.DeployCmd
	invoke *invokeCmd
	test   bool
	data   string
}

func (c *watchCmd) run() error {
	c.watch(func() {
		log.UI.Info("\nchanges detected - starting deploy")
		updated, err := c.deploy.Deploy()
		if err != nil {
			log.UI.Fatal(err)
		}
		if !updated {
			return
		}
		if c.invoke != nil {
			log.UI.Info("invoking function")
			if err := c.invoke.run(); err != nil {
				log.UI.Error(err)
			}
		}
		if c.test {
			log.UI.Info("running tests")
			err := shell.Exec(shell.ExecOptions{
				Args:    []string{"go", "test", "-v"},
				WorkDir: c.ctx.Path + "/test",
				Logger:  log.UI.Info,
			})
			if err != nil {
				log.UI.Error(err)
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
				log.UI.Fatal(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(c.ctx.Path); err != nil {
		log.UI.Fatal(err)
	}

	log.UI.Info("starting watch on go files in %s", c.ctx.Path)
	if err := w.Start(1 * time.Second); err != nil {
		log.UI.Fatal(err)
	}
}
