package workspace

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/mantil-io/mantil/cli/log"
	"gopkg.in/yaml.v2"
)

type FileStore struct {
	root string

	workspaceFile string
	projectRoot   string

	workspace *Workspace
	project   *Project
}

func (s *FileStore) restore() error {
	if err := s.loadWorkspace(); err != nil {
		if !errors.Is(err, ErrWorkspaceNotFound) {
			return err
		}
		s.workspace = newWorkspace(defaultWorkspaceName())
	}
	if s.projectRoot == "" {
		return nil
	}
	return s.loadProject()
}

func (s *FileStore) loadWorkspace() error {
	buf, err := ioutil.ReadFile(s.workspaceFile)
	if err != nil {
		if os.IsNotExist(err) {
			return log.Wrap(ErrWorkspaceNotFound)
		}
		return log.Wrap(err, "could not read workspace file")
	}
	var w Workspace
	if err = yaml.Unmarshal(buf, &w); err != nil {
		return log.Wrap(err, "could not unmarshal workspace")
	}
	w.afterRestore()
	s.workspace = &w
	return nil
}

func (s *FileStore) loadProject() error {
	buf, err := ioutil.ReadFile(configPath(s.projectRoot))
	if err != nil {
		return err
	}
	p := &Project{}
	if err := yaml.Unmarshal(buf, p); err != nil {
		return err
	}
	s.project = p
	s.linkProject()
	return nil
}

func (s *FileStore) linkProject() {
	s.project.workspace = s.workspace
	for _, stage := range s.project.Stages {
		stage.project = s.project
		// TODO ako ovo ne moze naci onda bum
		stage.account = s.workspace.Account(stage.Account)
		for _, f := range stage.Functions {
			f.stage = stage
		}
	}
}

func NewSingleDeveloperFileStore() (*FileStore, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, log.Wrap(err, "can't get user home dir")
	}
	usr, err := user.Current()
	if err != nil {
		return nil, log.Wrap(err, "can't find current user")
	}
	projectRoot, _ := FindProjectRoot(".")

	workspacePath := path.Join(home, ".mantil")
	if err := ensurePathExists(workspacePath); err != nil {
		return nil, err
	}

	w := &FileStore{
		workspaceFile: path.Join(workspacePath, usr.Username+".yml"),
		projectRoot:   projectRoot,
	}

	if err := w.restore(); err != nil {
		return nil, log.Wrap(err)
	}
	return w, err
}

func (s *FileStore) Workspace() *Workspace {
	return s.workspace
}

func (s *FileStore) Project() *Project {
	return s.project
}

func (s *FileStore) DefaultStage() *Stage {
	return s.Stage("")
}

func (s *FileStore) ProjectRoot() string {
	return s.projectRoot
}

func (s *FileStore) Stage(name string) *Stage {
	if s.project == nil {
		return nil
	}
	for _, s := range s.project.Stages {
		if s.Name == name || (name == "" && s.Default) {
			return s
		}
	}
	return nil
}

func (s *FileStore) Store() error {
	if s.project != nil {
		// TODO move SaveProject here
		if err := SaveProject(s.project, s.projectRoot); err != nil {
			return err
		}
	}
	return s.storeWorkspace()
}

func (s *FileStore) storeWorkspace() error {
	if s.workspace.Empty() {
		err := os.Remove(s.workspaceFile)
		if err != nil {
			return log.Wrap(err, "could not remove workspace")
		}
		return nil
	}
	buf, err := yaml.Marshal(s.workspace)
	if err != nil {
		return log.Wrap(err, "could not marshal workspace")
	}
	if err = ioutil.WriteFile(s.workspaceFile, buf, 0644); err != nil {
		return log.Wrap(err, "could not write workspace")
	}
	return nil
}

func ensurePathExists(dir string) error {
	err := os.Mkdir(dir, 0755)
	if os.IsExist(err) {
		return nil
	}
	return log.Wrap(err)
}
