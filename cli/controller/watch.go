package controller

import (
	"regexp"
	"time"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/radovskyb/watcher"
)

type WatchArgs struct {
	Method string
	Data   string
	Stage  string
	Test   bool
}

func Watch(a WatchArgs) error {
	fs, err := newStoreWithStage(a.Stage)
	if err != nil {
		return log.Wrap(err)
	}
	stage := fs.Stage(a.Stage)
	deploy, err := NewDeployWithStage(fs, stage)
	if err != nil {
		return log.Wrap(err)
	}
	var invoke func() error
	if a.Method != "" {
		invoke = InvokeCallback(stage, a.Method, a.Data, false, true)
	}
	var test func() error
	if a.Test {
		test = func() error {
			return runTests(fs.ProjectRoot(), stage.Endpoints.Rest, "")
		}
	}
	return runWatch(fs.ProjectRoot(), deploy, invoke, test)
}

func runWatch(path string, deploy *Deploy, invoke, test func() error) error {
	onChange := func() {
		ui.Info("")
		ui.Info("==> Changes detected")
		if err := deploy.Deploy(); err != nil {
			ui.Error(err)
			return
		}
		if !deploy.HasUpdates() {
			return
		}
		ui.Info("")
		if invoke != nil {
			ui.Info("==> Invoking function")
			if err := invoke(); err != nil {
				ui.Error(err)
			}
		}
		if test != nil {
			ui.Info("")
			ui.Info("==> Running tests")
			if err := test(); err != nil {
				ui.Error(err)
			}
		}
	}
	return runWatcher(onChange, path)
}

func runWatcher(onChange func(), path string) error {
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
				ui.Error(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(path); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Watching Go files in %s", path)
	if err := w.Start(1 * time.Second); err != nil {
		return log.Wrap(err)
	}
	return nil
}
