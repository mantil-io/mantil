package log

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mantil-io/mantil/domain"
)

func newEventsStore() *eventsStore {
	return &eventsStore{
		events: make(map[int64][]byte),
	}
}

func (s *eventsStore) push(buf []byte) {
	k := domain.NowMS()
	if _, ok := s.events[k]; ok {
		k++
	}
	s.events[k] = buf
}

type eventsStore struct {
	events map[int64][]byte
	dir    string
}

func (s *eventsStore) mkdir() error {
	appConfigDir, err := domain.AppConfigDir()
	if err != nil {
		return err
	}

	eventsDir := filepath.Join(appConfigDir, "events")
	if err := os.MkdirAll(eventsDir, os.ModePerm); err != nil {
		return Wrap(fmt.Errorf("failed to create events dir %s, error %w", eventsDir, err))
	}

	s.dir = eventsDir
	return nil
}

func (s *eventsStore) store() error {
	if err := s.mkdir(); err != nil {
		return err
	}
	for k, v := range s.events {
		path := filepath.Join(s.dir, fmt.Sprintf("%d", k))
		if err := os.WriteFile(path, v, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func (s *eventsStore) restore() error {
	if err := s.mkdir(); err != nil {
		return err
	}
	files, err := ioutil.ReadDir(s.dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		buf, err := ioutil.ReadFile(filepath.Join(s.dir, f.Name()))
		if err != nil {
			return err
		}
		k, err := strconv.Atoi(f.Name())
		if err == nil {
			s.events[int64(k)] = buf
		}
	}
	return nil
}

func (s *eventsStore) clear() error {
	if err := s.mkdir(); err != nil {
		return err
	}
	for k := range s.events {
		path := filepath.Join(s.dir, fmt.Sprintf("%d", k))
		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
		}
		delete(s.events, k)
	}
	return nil
}
