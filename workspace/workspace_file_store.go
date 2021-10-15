package workspace

import (
	"errors"
	"io/ioutil"
	"os"
	"path"

	"github.com/mantil-io/mantil/cli/log"
	"gopkg.in/yaml.v2"
)

type Store interface {
	Load(name string) (*Workspace, error)
	LoadOrNew(name string) (*Workspace, error)
	Save(*Workspace) error
}

type WorkspacesFileStore struct {
	root string
}

func NewSingleDeveloperWorkspacesFileStore() (*WorkspacesFileStore, error) {
	return newWorkspacesFileStore("")
}

func newWorkspacesFileStore(root string) (*WorkspacesFileStore, error) {
	s := WorkspacesFileStore{root: root}
	if root == "" {
		if err := s.setDefaultRoot(); err != nil {
			return nil, log.Wrap(err)
		}
	}
	if err := s.ensureRootExists(); err != nil {
		return nil, log.Wrap(err)
	}
	return &s, nil
}

func (s *WorkspacesFileStore) LoadOrNew(name string) (*Workspace, error) {
	if name == "" {
		name = defaultWorkspaceName()
	}
	ws, err := s.Load(name)
	if err == nil {
		return ws, nil
	}
	if errors.Is(err, ErrWorkspaceNotFound) {
		return newWorkspace(name), nil
	}
	return nil, err
}

func (s *WorkspacesFileStore) Load(name string) (*Workspace, error) {
	if name == "" {
		name = defaultWorkspaceName()
	}
	wsPath := s.workspacePath(name)
	buf, err := ioutil.ReadFile(wsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, log.Wrap(ErrWorkspaceNotFound)
		}
		return nil, log.Wrap(err, "could not read workspace file")
	}
	var w Workspace
	if err = yaml.Unmarshal(buf, &w); err != nil {
		return nil, log.Wrap(err, "could not unmarshal workspace")
	}
	w.afterRestore()
	return &w, nil
}

func (s *WorkspacesFileStore) workspacePath(name string) string {
	return path.Join(s.root, name+".yml")
}

func (s *WorkspacesFileStore) Save(w *Workspace) error {
	wsPath := s.workspacePath(w.Name)
	if w.Empty() {
		err := os.Remove(wsPath)
		if err != nil {
			return log.Wrap(err, "could not remove workspace config file")
		}
		return nil
	}
	buf, err := yaml.Marshal(w)
	if err != nil {
		return log.Wrap(err, "could not marshal workspace config")
	}
	if err = ioutil.WriteFile(wsPath, buf, 0644); err != nil {
		return log.Wrap(err, "could not write workspace config file")
	}
	return nil
}

func (s *WorkspacesFileStore) setDefaultRoot() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return log.Wrap(err)
	}
	s.root = path.Join(home, ".mantil")
	return nil
}

func (s *WorkspacesFileStore) ensureRootExists() error {
	return ensurePathExists(s.root)
}

func ensurePathExists(dir string) error {
	err := os.Mkdir(dir, 0755)
	if os.IsExist(err) {
		return nil
	}
	return log.Wrap(err)
}
