package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
)

func Env(stageName string) (string, *Stage) {
	initPath := "."
	path, err := FindProjectRoot(initPath)
	if err != nil {
		log.Fatal(err)
	}
	project, err := LoadProject(path)
	if err != nil {
		log.Fatal(err)
	}
	stage := project.Stage(stageName)
	var url string
	if stage != nil && stage.Endpoints != nil {
		url = stage.Endpoints.Rest
	}
	return fmt.Sprintf(`export %s='%s'
export %s='%s'
`, EnvProjectName, project.Name,
		EnvApiURL, url,
	), stage
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
