package workspace

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/mantil-io/mantil/cli/log"
	"gopkg.in/yaml.v2"
)

// enables setting workspace path outside of default
// can be used in test to don't mess with the default user workspace
const EnvWorkspacePath = "MANTIL_WORKSPACE_PATH"

type FileStore struct {
	workspaceFile string
	projectRoot   string
	workspace     *Workspace
	project       *Project
	environment   *EnvironmentConfig
}

func (s *FileStore) restore() error {
	if err := s.loadWorkspace(); err != nil {
		if !errors.Is(err, ErrWorkspaceNotFound) {
			return log.Wrap(err)
		}
		s.workspace = newWorkspace(defaultWorkspaceName())
	}
	if err := s.loadProject(); err != nil {
		return log.Wrap(err)
	}
	if err := s.loadEnvironment(); err != nil {
		return log.Wrap(err)
	}
	return factory(s.workspace, s.project, s.environment)
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
	s.workspace = &w
	return nil
}

func (s *FileStore) loadProject() error {
	if s.projectRoot == "" {
		return nil
	}
	buf, err := ioutil.ReadFile(configPath(s.projectRoot))
	if err != nil {
		return log.Wrap(err)
	}
	p := &Project{}
	if err := yaml.Unmarshal(buf, p); err != nil {
		return log.Wrap(err)
	}
	s.project = p
	return nil
}

func (s *FileStore) loadEnvironment() error {
	if s.projectRoot == "" {
		return nil
	}
	path := environmentConfigPath(s.projectRoot)
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return log.Wrap(err)
	}
	ec := &EnvironmentConfig{}
	if err := yaml.Unmarshal(buf, ec); err != nil {
		return log.Wrap(err)
	}
	s.environment = ec
	return nil
}

func defaultWorkspacePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", log.Wrap(err, "can't get user home dir")
	}
	workspacePath := path.Join(home, ".mantil")
	return workspacePath, nil
}

func workspacePath() (string, error) {
	if val, ok := os.LookupEnv(EnvWorkspacePath); ok {
		return val, nil
	}
	return defaultWorkspacePath()
}

func NewSingleDeveloperFileStore() (*FileStore, error) {
	workspacePath, err := workspacePath()
	if err != nil {
		return nil, log.Wrap(err)
	}
	if err := ensurePathExists(workspacePath); err != nil {
		return nil, err
	}
	projectRoot, _ := FindProjectRoot(".")
	w := &FileStore{
		workspaceFile: path.Join(workspacePath, defaultWorkspaceName()+".yml"),
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

// TODO: remove in new and then move inside file_store
func SaveProject(p *Project, basePath string) error {
	buf, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(basePath, configDir), os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(configPath(basePath), buf, 0644); err != nil {
		return err
	}
	return nil
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
