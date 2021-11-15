package domain

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/mantil-io/mantil/kit/schema"
	"github.com/pkg/errors"
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
			return errors.WithStack(err)
		}
		s.workspace = newWorkspace(defaultWorkspaceName())
	}
	if err := s.restoreState(); err != nil {
		return errors.WithStack(err)
	}
	if err := s.restoreEnvironment(); err != nil {
		return errors.WithStack(err)
	}
	return Factory(s.workspace, s.project, s.environment)
}

func (s *FileStore) restoreWorkspace() error {
	buf, err := ioutil.ReadFile(s.workspaceFile)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.WithStack(ErrWorkspaceNotFound)
		}
		return errors.Wrap(err, "could not read workspace file")
	}
	var w Workspace
	if err = yaml.Unmarshal(buf, &w); err != nil {
		return errors.Wrap(err, "could not unmarshal workspace")
	}
	s.workspace = &w
	return nil
}

func (s *FileStore) restoreState() error {
	if s.projectRoot == "" {
		return nil
	}
	p, err := readProjectState(s.projectRoot)
	if err != nil {
		return errors.WithStack(err)
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
		return errors.WithStack(err)
	}
	ec := &EnvironmentConfig{}
	schema, err := schema.From(ec)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := schema.ValidateYAML(buf); err != nil {
		return &EvironmentConfigValidationError{err}
	}
	if err := yaml.Unmarshal(buf, ec); err != nil {
		return errors.WithStack(err)
	}
	s.environment = ec
	return nil
}

func AppConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("can't read user config dir, error: %w", err)
	}
	appName := "Mantil" //os.Args[0]
	appConfigDir := filepath.Join(userConfigDir, appName)

	if err := os.MkdirAll(appConfigDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create application config dir %s, error %w", appConfigDir, err)
	}
	return appConfigDir, nil
}

const activationTokenFileName = ".token"

func StoreActivationToken(jwt string) error {
	dir, err := AppConfigDir()
	if err != nil {
		return err
	}
	filename := path.Join(dir, activationTokenFileName)
	return ioutil.WriteFile(filename, []byte(jwt), os.ModePerm)
}

func ReadActivationToken() (string, error) {
	dir, err := AppConfigDir()
	if err != nil {
		return "", err
	}
	filename := path.Join(dir, activationTokenFileName)
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("token not found")
		}
		return "", err
	}
	return string(buf), nil
}

func workspacePath() (string, error) {
	if val, ok := os.LookupEnv(EnvWorkspacePath); ok {
		return val, nil
	}
	return AppConfigDir()
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
		return nil, errors.WithStack(err)
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
		return nil, errors.WithStack(err)
	}
	w.syncProjects()
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
			return errors.Wrap(err, "could not remove workspace")
		}
		return nil
	}
	buf, err := yaml.Marshal(s.workspace)
	if err != nil {
		return errors.Wrap(err, "could not marshal workspace")
	}
	if err = ioutil.WriteFile(s.workspaceFile, buf, 0644); err != nil {
		return errors.Wrap(err, "could not write workspace")
	}
	return nil
}

func ensurePathExists(dir string) error {
	err := os.Mkdir(dir, 0755)
	if os.IsExist(err) {
		return nil
	}
	return errors.WithStack(err)
}

func (s *FileStore) NewProject(name, projectRoot string) error {
	if err := ValidateName(name); err != nil {
		return errors.WithStack(err)
	}
	project := &Project{
		Name: name,
	}
	if err := storeProject(project, projectRoot); err != nil {
		return errors.WithStack(err)
	}
	if err := createEnvironmentConfig(projectRoot); err != nil {
		return errors.WithStack(err)
	}
	s.workspace.AddProject(name, projectRoot)
	return s.Store()
}

func (s *FileStore) syncProjects() {
	if s.workspace == nil {
		return
	}
	var projects []*WorkspaceProject
	for _, p := range s.workspace.Projects {
		if !pathExists(p.Path) {
			continue
		}
		projects = append(projects, p)
	}
	s.workspace.Projects = projects
}

func (s *FileStore) NewStage(name, nodeName, projectName string) (*Stage, error) {
	st, err := s.project.NewStage(name, nodeName)
	if err != nil {
		return nil, err
	}
	st.node.AddStage(st.Name, projectName)
	return st, s.Store()
}

func (s *FileStore) RemoveStage(name string) error {
	st := s.project.Stage(name)
	st.node.RemoveStage(name)
	s.project.RemoveStage(name)
	return s.Store()
}

func (s *FileStore) GatherWorkspaceInfo() WorkspaceInfo {
	wi := WorkspaceInfo{
		Name:     s.workspace.Name,
		Nodes:    len(s.workspace.Nodes),
		Projects: len(s.workspace.Projects),
	}
	for _, pi := range s.workspace.Projects {
		p, err := readProjectState(pi.Path)
		if err != nil {
			continue
		}
		wi.Stages += p.NumberOfStages()
		wi.Functions += p.NumberOfFunctions()
	}
	return wi
}

func readProjectState(projectRoot string) (*Project, error) {
	buf, err := ioutil.ReadFile(stateFilePath(projectRoot))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	p := &Project{}
	if err := yaml.Unmarshal(buf, p); err != nil {
		return nil, errors.WithStack(err)
	}
	return p, nil
}

func createEnvironmentConfig(projectRoot string) error {
	path := environmentFilePath(projectRoot)
	if pathExists(path) {
		return nil
	}
	if err := ioutil.WriteFile(path, []byte(environmentConfigExample), 0644); err != nil {
		return errors.WithStack(err)
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
			return "", &ProjectNotFoundError{}
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

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err == nil {
		return true
	}
	return false
}
