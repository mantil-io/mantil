package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type BackendConfig struct {
	APIGatewayURL string `json:"apiGatewayURL"`
	Token         string `json:"token,omitempty"`
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

func BackendConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/.mantil/backend.json", home), nil
}

func LoadBackendConfig() (*BackendConfig, error) {
	path, err := BackendConfigPath()
	if err != nil {
		return nil, fmt.Errorf("could not get backend config path - %v", err)
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read backend config file - %v", err)
	}
	config := &BackendConfig{}
	if err = json.Unmarshal(buf, config); err != nil {
		return nil, fmt.Errorf("could not unmarshal backend config - %v", err)
	}
	return config, nil
}

func (bc *BackendConfig) Save() error {
	path, err := BackendConfigPath()
	if err != nil {
		return fmt.Errorf("could not get backend config path - %v", err)
	}
	buf, err := json.Marshal(bc)
	if err != nil {
		return fmt.Errorf("could not marshal backend config - %v", err)
	}
	if err = ioutil.WriteFile(path, buf, 0644); err != nil {
		return fmt.Errorf("could not write backend config file - %v", err)
	}
	return nil
}

func BackendURL() (string, error) {
	if url := os.Getenv("MANTIL_BACKEND_URL"); url != "" {
		return url, nil
	}
	config, err := LoadBackendConfig()
	if err != nil {
		return "", fmt.Errorf("could not load backend config - %v", err)
	}
	return config.APIGatewayURL, nil
}
