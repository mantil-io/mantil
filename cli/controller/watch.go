package controller

import (
	"os"
	"os/signal"
	"regexp"
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
	deploy, err := NewDeployWithStage(fs, stage)
	if err != nil {
		return log.Wrap(err)
	}
	w := watch{
		deploy: func() (bool, error) {
			if err := deploy.DeployWithTitle("Changes spotted! Starting deploy"); err != nil {
				return false, err
			}
			return deploy.HasUpdates(), nil
		},
	}

	if a.Method != "" {
		w.invoke = stageInvokeCallback(stage, a.Method, a.Data, true, buildShowResponseHandler(false))
	}
	if a.Test {
		w.test = func() error {
			return runTests(fs.ProjectRoot(), stage.Endpoints.Rest, "")
		}
	}

	return w.run(fs.ProjectRoot())
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

func (w *watch) run(path string) error {
	wr := watcher.New()
	wr.SetMaxEvents(1)
	wr.FilterOps(watcher.Write, watcher.Create, watcher.Remove)

	// only watch for changes in go files
	r := regexp.MustCompile(`\.go$`)
	wr.AddFilterHook(watcher.RegexFilterHook(r, false))

	ctrlc := make(chan os.Signal, 1)
	signal.Notify(ctrlc, syscall.SIGINT)
	go func() {
		for {
			select {
			case <-wr.Event:
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
