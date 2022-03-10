package controller

import (
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/radovskyb/watcher"
)

type WatchArgs struct {
	Method string
	Data   string
	Stage  string
	Test   bool
}

func Watch(a WatchArgs) error {
	fs, stage, err := newStoreWithStage(a.Stage)
	if err != nil {
		return log.Wrap(err)
	}
	w := watch{
		deploy: func() (bool, error) {
			fs, stage, err = newStoreWithStage(a.Stage)
			if err != nil {
				return false, log.Wrap(err)
			}
			deploy, err := NewDeployWithStage(fs, stage)
			if err != nil {
				return false, log.Wrap(err)
			}
			if err := deploy.DeployWithTitle("Changes spotted! Starting deploy"); err != nil {
				return false, log.Wrap(err)
			}
			return deploy.HasUpdates(), nil
		},
	}

	if a.Method != "" {
		invoke, err := stageInvokeCallback(stage, a.Method, a.Data, true, buildShowResponseHandler(false))
		if err != nil {
			return log.Wrap(err)
		}
		w.invoke = invoke
	}
	if a.Test {
		w.test = func() error {
			return runTests(fs.ProjectRoot(), stage.RestEndpoint(), "")
		}
	}

	// add separator at the end so dirs with prefix BuildDir are not matched
	buildDirPath := filepath.Join(fs.ProjectRoot(), BuildDir) + string(filepath.Separator)
	return w.run(fs.ProjectRoot(), []string{buildDirPath})
}

type watch struct {
	deploy func() (bool, error)
	invoke func() error
	test   func() error
	cycles int
}

func (w *watch) onChange() {
	w.cycles++
	var hasUpdates bool
	var err error
	tmr := timerFn()
	defer func() {
		log.Event(domain.Event{WatchCycle: &domain.WatchCycle{
			Duration:   tmr(),
			CycleNo:    w.cycles,
			HasUpdates: hasUpdates,
			Invoke:     w.invoke != nil,
			Test:       w.test != nil,
		}})
		_ = log.SendEvents()
	}()

	hasUpdates, err = w.deploy()
	if err != nil {
		ui.Error(err)
		return
	}
	if !hasUpdates {
		return
	}
	if w.invoke != nil {
		ui.Info("")
		ui.Info("Invoking function...")
		if err = w.invoke(); err != nil {
			ui.Error(err)
		}
	}
	if w.test != nil {
		ui.Info("")
		ui.Info("Running tests...")
		if err = w.test(); err != nil {
			ui.Error(err)
		}
	}
}

func (w *watch) run(path string, ignoredDirs []string) error {
	wr := watcher.New()
	wr.SetMaxEvents(1)
	wr.FilterOps(watcher.Write, watcher.Create, watcher.Remove)

	// only watch for changes in go files
	r := regexp.MustCompile(`\.go$`)
	wr.AddFilterHook(watcher.RegexFilterHook(r, false))

	isIgnoredPath := func(path string) bool {
		for _, dir := range ignoredDirs {
			if strings.HasPrefix(path, dir) {
				return true
			}
		}
		return false
	}

	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, syscall.SIGINT)
	go func() {
		for {
			select {
			case e := <-wr.Event:
				// due to contraints of the current watcher library finding a way to ignore automatically generated
				// build folder with other filters proved to be a challenge, adding this workaround for now.
				if isIgnoredPath(e.Path) {
					continue
				}
				w.onChange()
				ui.Info("")
				ui.Info("Watching changes in %s", path)
			case err := <-wr.Error:
				ui.Error(err)
			case <-wr.Closed:
				return
			case <-ctrlc:
				wr.Close()
			}
		}
	}()

	if err := wr.AddRecursive(path); err != nil {
		return log.Wrap(err)
	}
	ui.Info("")
	ui.Info("Watching changes in %s", path)
	if err := wr.Start(1 * time.Second); err != nil {
		return log.Wrap(err)
	}

	log.Event(domain.Event{WatchDone: &domain.WatchDone{Cycles: w.cycles}})
	return nil
}
