package domain

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// enables setting workspace path outside of default
// can be used in test to don't mess with the default user workspace
const EnvWorkspacePath = "MANTIL_WORKSPACE_PATH"

const (
	configDir               = "config"
	configFilename          = "state.yml"
	environmentFilename     = "environment.yml"
	workspaceFilename       = "workspace.yml"
	activationTokenFilename = ".token"
)

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
		s.workspace = newWorkspace()
		_ = s.storeWorkspace()
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
	s.workspace.upgrade() // TODO safe to remove after 0.2 release
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
	ec, err := ValidateEnvironmentConfig(buf)
	if err != nil {
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

func StoreActivationToken(jwt string) error {
	dir, _, err := WorkspacePathAndName()
	if err != nil {
		return err
	}
	return StoreActivationTokenTo(jwt, dir)
}

func StoreActivationTokenTo(jwt string, dir string) error {
	filename := path.Join(dir, activationTokenFilename)
	return ioutil.WriteFile(filename, []byte(jwt), os.ModePerm)
}

func ReadActivationToken() (string, error) {
	dir, _, err := WorkspacePathAndName()
	if err != nil {
		return "", err
	}
	filename := path.Join(dir, activationTokenFilename)
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.WithStack(fmt.Errorf("token not found"))
		}
		return "", err
	}
	return string(buf), nil
}

func WorkspacePathAndName() (string, string, error) {
	// workspace pats set from env, used in end_to_end test
	if val, ok := os.LookupEnv(EnvWorkspacePath); ok {
		return val, workspaceFilename, nil
	}
	// if we are in the project and exists config/workspace.yml use that
	projectRoot, err := FindProjectRoot(".")
	if err == nil {
		if pathExists(filepath.Join(projectRoot, configDir, workspaceFilename)) {
			return filepath.Join(projectRoot, configDir), workspaceFilename, nil
		}
	}
	apd, err := AppConfigDir()
	if err != nil {
		return "", "", err
	}
	path := filepath.Join(apd, workspaceFilename)
	if pathExists(path) {
		return apd, workspaceFilename, nil
	}
	legacyPath := filepath.Join(apd, legacyWorkspaceName()+".yml")
	if pathExists(legacyPath) {
		upgradeWorkspace(legacyPath, path)
	}
	return apd, workspaceFilename, nil
}

func upgradeWorkspace(from, to string) {
	var s FileStore
	s.workspaceFile = from
	_ = s.restoreWorkspace()
	s.workspace.ID = UID()
	s.workspace.CreatedAt = time.Now().UnixMilli()
	s.workspaceFile = to
	_ = s.storeWorkspace()
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
	projectRoot, err := FindProjectRoot(".")
	if err != nil && mustFindProject {
		return nil, err
	}
	workspacePath, workspaceFilename, err := WorkspacePathAndName()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := ensurePathExists(workspacePath); err != nil {
		return nil, err
	}
	w := &FileStore{
		workspaceFile: filepath.Join(workspacePath, workspaceFilename),
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

func (s *FileStore) marshalWorkspace() ([]byte, error) {
	s.workspace.Version = Version() // store last version which update workspace file
	buf, err := yaml.Marshal(s.workspace)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal workspace")
	}
	return buf, nil
}

func (s *FileStore) storeWorkspace() error {
	buf, err := s.marshalWorkspace()
	if err != nil {
		return err
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
	return nil
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

func (s *FileStore) AsCliWorkspace() *CliWorkspace {
	if s == nil {
		return nil
	}
	wi := CliWorkspace{
		ID:       s.workspace.ID,
		Nodes:    len(s.workspace.Nodes),
		Projects: len(s.workspace.Projects),
	}
	for _, pi := range s.workspace.Projects {
		p, err := readProjectState(pi.Path)
		if err != nil {
			continue
		}
		wi.Stages += p.NumberOfStages()
		wi.Functions += p.numberOfFunctions()
	}

	m := make(map[string]struct{})
	for _, n := range s.workspace.Nodes {
		m[n.AccountID] = struct{}{}
	}
	wi.AWSAccounts = len(m)

	m = make(map[string]struct{})
	for _, n := range s.workspace.Nodes {
		m[n.Region] = struct{}{}
	}
	wi.AWSRegions = len(m)
	return &wi
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
	return filepath.Join(projectRoot, configDir, environmentFilename)
}

func stateFilePath(projectRoot string) string {
	return filepath.Join(projectRoot, configDir, configFilename)
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
