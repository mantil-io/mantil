package watch

import (
	"log"
	"regexp"
	"time"

	"github.com/radovskyb/watcher"
)

func Start(path string, onChange func()) {
	w := watcher.New()

	w.SetMaxEvents(1)
	w.FilterOps(watcher.Write)

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

	if err := w.AddRecursive(path); err != nil {
		log.Fatal(err)
	}

	log.Printf("starting watch for all go files in %s", path)
	if err := w.Start(1 * time.Second); err != nil {
		log.Fatal(err)
	}
}
