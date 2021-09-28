package cmd

import (
	"regexp"
	"time"

	"github.com/mantil-io/mantil.go/pkg/shell"
	"github.com/mantil-io/mantil/cli/commands/deploy"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/radovskyb/watcher"
)

type watchCmd struct {
	repoPath string
	deploy   *deploy.DeployCmd
	invoke   *invokeCmd
	test     bool
	data     string
}

func (c *watchCmd) run() error {
	c.watch(func() {
		log.Info("\nchanges detected - starting deploy")
		updated, err := c.deploy.Deploy()
		if err != nil {
			log.Fatal(err)
		}
		if !updated {
			return
		}
		if c.invoke != nil {
			log.Info("invoking function")
			if err := c.invoke.run(); err != nil {
				log.Error(err)
			}
		}
		if c.test {
			log.Info("running tests")
			err := shell.Exec(shell.ExecOptions{
				Args:    []string{"go", "test", "-v"},
				WorkDir: c.repoPath + "/test",
				Logger:  log.Info,
			})
			if err != nil {
				log.Error(err)
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
				log.Fatal(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(c.repoPath); err != nil {
		log.Fatal(err)
	}

	log.Info("starting watch on go files in %s", c.repoPath)
	if err := w.Start(1 * time.Second); err != nil {
		log.Fatal(err)
	}
}
