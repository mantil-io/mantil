package commands

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

type WorkspaceConfig struct {
	Name     string           `yaml:"name"`
	Accounts []*AccountConfig `yaml:"accounts"`
}

type AccountConfig struct {
	Name      string            `yaml:"name"`
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

func (w *WorkspaceConfig) DefaultAccount() *AccountConfig {
	for _, a := range w.Accounts {
		if a.Name == DefaultAccountName {
			return a
		}
	}
	return nil
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

func RemoveConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return os.RemoveAll(fmt.Sprintf("%s/.mantil", home))
}

func WorkspaceConfigPath() (string, error) {
	name := DefaultWorkspaceName()
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.mantil/%s.yml", home, name), nil
}

func DefaultWorkspaceName() string {
	u, _ := user.Current()
	if u == nil {
		return ""
	}
	return strings.ToLower(u.Name)
}

func LoadWorkspaceConfig() (*WorkspaceConfig, error) {
	path, err := WorkspaceConfigPath()
	if err != nil {
		return nil, fmt.Errorf("could not get workspace config path - %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &WorkspaceConfig{
			Name: DefaultWorkspaceName(),
		}, nil
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read workspace config file - %v", err)
	}
	config := &WorkspaceConfig{}
	if err = yaml.Unmarshal(buf, config); err != nil {
		return nil, fmt.Errorf("could not unmarshal workspace config - %v", err)
	}
	return config, nil
}

func (wc *WorkspaceConfig) Save() error {
	path, err := WorkspaceConfigPath()
	if err != nil {
		return fmt.Errorf("could not get workspace config path - %v", err)
	}
	buf, err := yaml.Marshal(wc)
	if err != nil {
		return fmt.Errorf("could not marshal workspace config - %v", err)
	}
	if err = ioutil.WriteFile(path, buf, 0644); err != nil {
		return fmt.Errorf("could not write workspace config file - %v", err)
	}
	return nil
}

func (wc *WorkspaceConfig) Account(name string) *AccountConfig {
	for _, a := range wc.Accounts {
		if a.Name == name {
			return a
		}
	}
	return nil
}

func RestEndpoint(accountName string) (string, error) {
	config, err := LoadWorkspaceConfig()
	if err != nil {
		return "", err
	}
	account := config.Account(accountName)
	if account == nil {
		return "", fmt.Errorf("account not found")
	}
	return account.Endpoints.Rest, nil
}

func WsEndpoint(accountName string) (string, error) {
	config, err := LoadWorkspaceConfig()
	if err != nil {
		return "", err
	}
	account := config.Account(accountName)
	if account == nil {
		return "", fmt.Errorf("account not found")
	}
	return fmt.Sprintf("%s/$default", account.Endpoints.Ws), nil
}

func DefaultRestEndpoint() (string, error) {
	return RestEndpoint(DefaultAccountName)
}

func DefaultWsEndpoint() (string, error) {
	return WsEndpoint(DefaultAccountName)
}
