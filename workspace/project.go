package workspace

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	configDir  = "config"
	configName = "project.yml"
)

const (
	EnvProjectName = "MANTIL_PROJECT_NAME"
	EnvStageName   = "MANTIL_STAGE_NAME"
	EnvApiURL      = "MANTIL_API_URL"
)

const (
	TagProjectName = "MANTIL_PROJECT"
)

type Project struct {
	Name   string   `yaml:"name"` // required
	Stages []*Stage `yaml:"stages,omitempty"`
}

func (p *Project) ResourceTags() map[string]string {
	return map[string]string{
		TagProjectName: p.Name,
	}
}

func SaveProject(p *Project, basePath string) error {
	buf, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(basePath, configDir), os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(configPath(basePath), buf, 0644); err != nil {
		return err
	}
	return nil
}

func LoadProject(basePath string) (*Project, error) {
	buf, err := ioutil.ReadFile(configPath(basePath))
	if err != nil {
		return nil, err
	}
	p := &Project{}
	if err := yaml.Unmarshal(buf, p); err != nil {
		return nil, err
	}
	return p, nil
}

func configPath(basePath string) string {
	return filepath.Join(basePath, configDir, configName)
}

func (p *Project) Stage(name string) *Stage {
	for _, s := range p.Stages {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (p *Project) DefaultStage() *Stage {
	for _, s := range p.Stages {
		if s.Default {
			return s
		}
	}
	return nil
}

func (s *Stage) SetPublicBucket(name, bucket string) bool {
	for _, p := range s.Public {
		if p.Name == name {
			p.Bucket = bucket
			return true
		}
	}
	return false
}

func (s *Stage) SetEndpoints(rest, ws string) {
	// TODO: zasto je ovo pointer
	s.Endpoints = &StageEndpoints{
		Rest: rest,
		Ws:   ws,
	}
}

func (p *Project) SetDefaultStage() {
	if len(p.Stages) == 0 {
		return
	}
	if s := p.DefaultStage(); s != nil {
		return
	}
	if s := p.Stage(DefaultStageName); s != nil {
		s.Default = true
		return
	}
	p.Stages[0].Default = true
}

func (p *Project) UpsertStage(stage *Stage) {
	for idx, s := range p.Stages {
		if s.Name == stage.Name {
			p.Stages[idx] = stage
			return
		}
	}
	p.Stages = append(p.Stages, stage)
}

func (p *Project) RemoveStage(stageName string) {
	for idx, s := range p.Stages {
		if s.Name == stageName {
			p.Stages = append(p.Stages[:idx], p.Stages[idx+1:]...)
		}
	}
}

func ProjectResource(projectName string, stageName string, v ...string) string {
	r := fmt.Sprintf("mantil-project-%s-%s", projectName, stageName)
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}

func FindProjectRoot(initialPath string) (string, error) {
	currentPath := initialPath
	for {
		_, err := os.Stat(filepath.Join(currentPath, configPath(initialPath)))
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
			return "", fmt.Errorf("no mantil project found")
		}
		currentPath += "/.."
	}
}

func (s *Stage) AddFunctionDefaults() {
	for _, f := range s.Functions {
		if f.Path == "" {
			f.Path = f.Name
		}
		if f.S3Key == "" {
			if f.Hash != "" {
				f.S3Key = fmt.Sprintf("functions/%s-%s.zip", f.Name, f.Hash)
			} else {
				f.S3Key = fmt.Sprintf("functions/%s.zip", f.Name)
			}
		}
		if f.Runtime == "" {
			f.Runtime = "provided.al2"
		}
		if f.MemorySize == 0 {
			f.MemorySize = 128
		}
		if f.Timeout == 0 {
			f.Timeout = 60 * 15
		}
		if f.Handler == "" {
			f.Handler = "bootstrap"
		}
		if f.Env == nil {
			f.Env = make(map[string]string)
		}
	}
}
