package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFile = "mantil.yaml"
)

type Config struct {
	Functions map[string]Function `yaml:"functions"`
}

type Function struct {
	Path string `yaml:"path"`
}

func LoadConfig() (*Config, error) {
	yf, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return nil, err
	}

	c := Config{
		Functions: make(map[string]Function),
	}
	err = yaml.Unmarshal(yf, c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func Exists() bool {
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		return false
	}
	return true
}
