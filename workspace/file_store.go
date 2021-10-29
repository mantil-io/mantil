package workspace

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/kit/schema"
	"gopkg.in/yaml.v2"
)

// enables setting workspace path outside of default
// can be used in test to don't mess with the default user workspace
const EnvWorkspacePath = "MANTIL_WORKSPACE_PATH"

const stateFileHeader = `# DO NOT EDIT.
# This is Mantil project state file maintained by Mantil.
# It is stored in human readable format so you can
# add it to git and have history of all changes.
`

type FileStore struct {
	workspaceFile string
	projectRoot   string
	workspace     *Workspace
	project       *Project
	environment   *EnvironmentConfig
}

func (s *FileStore) restore() error {
	if err := s.restoreWorkspace(); err != nil {
		if !errors.Is(err, ErrWorkspaceNotFound) {
			return log.Wrap(err)
		}
		s.workspace = newWorkspace(defaultWorkspaceName())
	}
	if err := s.restoreState(); err != nil {
		return log.Wrap(err)
	}
	if err := s.restoreEnvironment(); err != nil {
		return log.Wrap(err)
	}
	return Factory(s.workspace, s.project, s.environment)
}

func (s *FileStore) restoreWorkspace() error {
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

func (s *FileStore) restoreState() error {
	if s.projectRoot == "" {
		return nil
	}
	buf, err := ioutil.ReadFile(stateFilePath(s.projectRoot))
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

func (s *FileStore) restoreEnvironment() error {
	if s.projectRoot == "" {
		return nil
	}
	path := environmentFilePath(s.projectRoot)
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return log.Wrap(err)
	}
	ec := &EnvironmentConfig{}
	schema, err := schema.From(ec)
	if err != nil {
		return log.Wrap(err)
	}
	if err := schema.ValidateYAML(buf); err != nil {
		return &EvironmentConfigValidationError{err}
	}
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

// NewSingleDeveloperWorkspaceStore loads workspace
// allows to be outside of project
func NewSingleDeveloperWorkspaceStore() (*FileStore, error) {
	return newSingleDeveloper(false)
}

// NewSingleDeveloperProject loads workspace and project config files
func NewSingleDeveloperProjectStore() (*FileStore, error) {
	return newSingleDeveloper(true)
}

func newSingleDeveloper(mustFindProject bool) (*FileStore, error) {
	workspacePath, err := workspacePath()
	if err != nil {
		return nil, log.Wrap(err)
	}
	if err := ensurePathExists(workspacePath); err != nil {
		return nil, err
	}
	projectRoot, err := FindProjectRoot(".")
	if err != nil && mustFindProject {
		return nil, err
	}
	w := &FileStore{
		workspaceFile: path.Join(workspacePath, defaultWorkspaceName()+".yml"),
		projectRoot:   projectRoot,
	}
	if err := w.restore(); err != nil {
		return nil, log.Wrap(err)
	}
	return w, nil
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
		if err := storeProject(s.project, s.projectRoot); err != nil {
			return err
		}
	}
	return s.storeWorkspace()
}

func storeProject(p *Project, projectRoot string) error {
	buf, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(projectRoot, configDir), os.ModePerm); err != nil {
		return err
	}
	buf = append([]byte(stateFileHeader), buf...)
	if err := ioutil.WriteFile(stateFilePath(projectRoot), buf, 0644); err != nil {
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

func (s *FileStore) NewProject(name, projectRoot string) error {
	if err := ValidateName(name); err != nil {
		return log.Wrap(err)
	}
	project := &Project{
		Name: name,
	}
	if err := storeProject(project, projectRoot); err != nil {
		return log.Wrap(err)
	}
	if err := createEnvironmentConfig(projectRoot); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func createEnvironmentConfig(projectRoot string) error {
	path := environmentFilePath(projectRoot)
	if fileExists(path) {
		return nil
	}
	if err := ioutil.WriteFile(path, []byte(environmentConfigExample), 0644); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func FindProjectRoot(initialPath string) (string, error) {
	currentPath := initialPath
	for {
		_, err := os.Stat(filepath.Join(currentPath, stateFilePath(initialPath)))
		if err == nil {
			abs, err := filepath.Abs(currentPath)
			if err != nil {
				return "", err
			}
			return abs, nil
		}
		currentPathAbs, err := filepath.Abs(currentPath)
		if err != nil {
			return "", err
		}
		if currentPathAbs == "/" {
			return "", ErrProjectNotFound
		}
		currentPath += "/.."
	}
}

func environmentFilePath(projectRoot string) string {
	return filepath.Join(projectRoot, configDir, environmentConfigName)
}

func stateFilePath(projectRoot string) string {
	return filepath.Join(projectRoot, configDir, configName)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err == nil {
		return true
	}
	return false
}
