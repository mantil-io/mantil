package mantil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

const (
	localConfigPath = "config/mantil.local.json"
)

type LocalProjectConfig struct {
	Name   string `json:"name"`
	ApiURL string `json:"apiURL,omitempty"`
}

func CreateLocalConfig(name string) (*LocalProjectConfig, error) {
	lc := LocalConfig(name)
	return lc, lc.Save(name)
}

func LocalConfig(name string) *LocalProjectConfig {
	return &LocalProjectConfig{
		Name: name,
	}
}

func (c *LocalProjectConfig) Save(path string) error {
	buf, err := json.Marshal(c)
	if err != nil {
		return err
	}
	configDir := filepath.Join(path, "config")
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(path, localConfigPath), buf, 0644); err != nil {
		return err
	}
	return nil
}

func LoadLocalConfig(projectRoot string) (*LocalProjectConfig, error) {
	buf, err := ioutil.ReadFile(filepath.Join(projectRoot, localConfigPath))
	if err != nil {
		return nil, err
	}
	c := &LocalProjectConfig{}
	if err := json.Unmarshal(buf, c); err != nil {
		return nil, err
	}
	return c, nil
}

func Env() (string, *LocalProjectConfig) {
	initPath := "."
	path, err := FindProjectRoot(initPath)
	if err != nil {
		log.Fatal(err)
	}
	config, err := LoadLocalConfig(path)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf(`export %s='%s'
export %s='%s'
`, EnvProjectName, config.Name,
		EnvApiURL, config.ApiURL,
	), config
}

func SaveToken(projectName, token string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := path.Join(home, ".mantil", projectName)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	config := path.Join(configDir, "config")
	if err := ioutil.WriteFile(config, []byte(token), 0755); err != nil {
		return err
	}
	return nil
}

func ReadToken(projectName string) (string, error) {
	token := os.Getenv("MANTIL_TOKEN")
	if token != "" {
		return token, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	config := path.Join(home, ".mantil", projectName, "config")
	data, err := ioutil.ReadFile(config)
	if err != nil {
		return "", err
	}
	token = string(data)
	if token == "" {
		return "", fmt.Errorf("token not found")
	}
	return token, nil
}
