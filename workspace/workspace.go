package workspace

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/cli/log"
	"gopkg.in/yaml.v2"
)

const (
	DefaultAccountName = "dev"
)

const (
	TagWorkspace = "MANTIL_WORKSPACE"
	TagKey       = "MANTIL_KEY"
)

const (
	EnvWorkspaceKey = "MANTIL_WORKSPACE_KEY"
)

type Workspace struct {
	Name     string     `yaml:"name"`
	UID      string     `yaml:"uid"`
	Accounts []*Account `yaml:"accounts"`
}

type Account struct {
	Name      string           `yaml:"name"`
	ID        string           `yaml:"id"`
	Region    string           `yaml:"region"`
	Bucket    string           `yaml:"bucket"`
	Keys      AccountKeys      `yaml:"keys"`
	Endpoints AccountEndpoints `yaml:"endpoints"`
	Functions AccountFunctions `yaml:"functions"`
	CliRole   string           `yaml:"cli_role"`
	workspace *Workspace
}

type AccountKeys struct {
	Public  string `yaml:"public"`
	Private string `yaml:"private"`
}

type AccountEndpoints struct {
	Rest string `yaml:"rest"`
	Ws   string `yaml:"ws"`
}

type AccountFunctions struct {
	Bucket string `yaml:"bucket"`
	Path   string `yaml:"key"`
}

func (c *Account) Marshal() ([]byte, error) {
	return yaml.Marshal(c)
}

func (w *Workspace) DefaultAccount() *Account {
	for _, a := range w.Accounts {
		if a.Name == DefaultAccountName {
			return a
		}
	}
	return nil
}

func (w *Workspace) UpsertAccount(ac *Account) {
	for _, a := range w.Accounts {
		if a.Name == ac.Name {
			*a = *ac
			return
		}
	}
	w.Accounts = append(w.Accounts, ac)
}

func (w *Workspace) RemoveAccount(name string) {
	for idx, a := range w.Accounts {
		if a.Name == name {
			w.Accounts = append(w.Accounts[:idx], w.Accounts[idx+1:]...)
			return
		}
	}
}

func createConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	err = os.Mkdir(fmt.Sprintf("%s/.mantil", home), 0755)
	if os.IsExist(err) {
		return nil
	}
	return err
}

func workspaceConfigPath() (string, error) {
	name := defaultWorkspaceName()
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.mantil/%s.yml", home, name), nil
}

func defaultWorkspaceName() string {
	u, _ := user.Current()
	if u == nil {
		return ""
	}
	return strings.ToLower(u.Username)
}

func Load() (*Workspace, error) {
	path, err := workspaceConfigPath()
	if err != nil {
		return nil, fmt.Errorf("could not get workspace config path - %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Workspace{
			Name: defaultWorkspaceName(),
		}, nil
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read workspace config file - %v", err)
	}
	config := &Workspace{}
	if err = yaml.Unmarshal(buf, config); err != nil {
		return nil, fmt.Errorf("could not unmarshal workspace config - %v", err)
	}
	return config, nil
}

func UpsertAccount(ac *Account) error {
	config, err := Load()
	if err != nil {
		return fmt.Errorf("could not load workspace config - %v", err)
	}
	config.UpsertAccount(ac)
	if err := createConfigDir(); err != nil {
		return fmt.Errorf("could not create config directory - %v", err)
	}
	if err := config.Save(); err != nil {
		return fmt.Errorf("could not save backend config - %v", err)
	}
	return nil
}

func RemoveAccount(name string) error {
	config, err := Load()
	if err != nil {
		return err
	}
	config.RemoveAccount(name)
	if err := config.Save(); err != nil {
		return err
	}
	return nil
}

func (w *Workspace) Save() error {
	path, err := workspaceConfigPath()
	if err != nil {
		return fmt.Errorf("could not get workspace config path - %v", err)
	}
	buf, err := yaml.Marshal(w)
	if err != nil {
		return fmt.Errorf("could not marshal workspace config - %v", err)
	}
	if err = ioutil.WriteFile(path, buf, 0644); err != nil {
		return fmt.Errorf("could not write workspace config file - %v", err)
	}
	return nil
}

func (w *Workspace) Account(name string) *Account {
	for _, a := range w.Accounts {
		if a.Name == name {
			return a
		}
	}
	return nil
}

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

var (
	ErrAccountExists     = fmt.Errorf("account already exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
)

func (w *Workspace) NewAccount(name, awsAccountID, awsRegion, functionsBucket, functionsPath string) (*Account, error) {
	if w.accountExists(name) {
		return nil, ErrAccountExists
	}
	publicKey, privateKey, err := auth.CreateKeyPair()
	if err != nil {
		return nil, log.Wrap(err, "could not create public/private key pair")
	}
	a := &Account{
		Name:   name,
		ID:     awsAccountID,
		Region: awsRegion,
		Bucket: bucket(awsRegion, awsAccountID),
		Keys: AccountKeys{
			Public:  publicKey,
			Private: privateKey,
		},
		Functions: AccountFunctions{
			Bucket: functionsBucket,
			Path:   functionsPath,
		},
		workspace: w,
	}
	w.Accounts = append(w.Accounts, a)
	return a, nil
}

func (w *Workspace) accountExists(name string) bool {
	for _, a := range w.Accounts {
		if a.Name == name {
			return true
		}
	}
	return false
}

func newWorkspace(name string) *Workspace {
	return &Workspace{
		Name: name,
		UID:  uid(),
	}
}

func (w *Workspace) ResourceName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, w.UID)
}

func (w *Workspace) ResourceTags() map[string]string {
	return map[string]string{
		TagWorkspace: w.Name,
		TagKey:       w.UID,
	}
}

func (w *Workspace) Empty() bool {
	return len(w.Accounts) == 0
}

// idea stolen from:  https://github.com/nats-io/nats-server/blob/fd9e9480dad9498ed8109e659fc8ed5c9b2a1b41/server/nkey.go#L41
func uid() string {
	var rndData [4]byte
	data := rndData[:]
	_, _ = io.ReadFull(rand.Reader, data)
	var encoded [7]byte
	base32.StdEncoding.WithPadding(base32.NoPadding).Encode(encoded[:], data)
	return strings.ToLower(string(encoded[:]))
}

func (a *Account) ResourceSuffix() string {
	return a.workspace.UID
}

func (w *Workspace) afterRestore() {
	for _, a := range w.Accounts {
		a.workspace = w
	}
	// when restored from previous version
	if w.UID == "" && len(w.Accounts) == 0 {
		w.UID = uid()
	}
}
