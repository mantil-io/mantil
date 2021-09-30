package workspace

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	DefaultAccountName = "dev"
)

type Workspace struct {
	Name     string     `yaml:"name"`
	Accounts []*Account `yaml:"accounts"`
}

type Account struct {
	Name      string            `yaml:"name"`
	Bucket    string            `yaml:"bucket"`
	Keys      *AccountKeys      `yaml:"keys"`
	Endpoints *AccountEndpoints `yaml:"endpoints"`
}

type AccountKeys struct {
	Public  string `yaml:"public"`
	Private string `yaml:"private"`
}

type AccountEndpoints struct {
	Rest string `yaml:"rest"`
	Ws   string `yaml:"ws"`
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

func CreateConfigDir() error {
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
	if err := CreateConfigDir(); err != nil {
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
