package workspace

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/cli/log"
	"gopkg.in/yaml.v2"
)

const (
	configDir             = "config"
	configName            = "project.yml"
	environmentConfigName = "environment.yml"
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

func (s *Stage) SetPublicBucket(bucket string) {
	if s.Public == nil {
		s.Public = &Public{}
	}
	s.Public.Bucket = bucket
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
	r := fmt.Sprintf("%s-%s", projectName, stageName)
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
		f.addDefaults()
	}
}

func (s *Stage) AddFunction(name string) {
	f := &Function{
		Name: name,
	}
	f.addDefaults()
	s.Functions = append(s.Functions, f)
}

func (s *Stage) RemoveFunctions(removed []string) {
	for _, r := range removed {
		for idx, sf := range s.Functions {
			if sf.Name == r {
				s.Functions = append(s.Functions[:idx], s.Functions[idx+1:]...)
				break
			}
		}
	}
}

type EnvironmentConfig struct {
	Project ProjectEnvironmentConfig `yaml:"project"`
}

type ProjectEnvironmentConfig struct {
	Env    map[string]string        `yaml:"env"`
	Stages []StageEnvironmentConfig `yaml:"stages"`
}

type StageEnvironmentConfig struct {
	Name      string                      `yaml:"name"`
	Env       map[string]string           `yaml:"env"`
	Functions []FunctionEnvironmentConfig `yaml:"functions"`
}

type FunctionEnvironmentConfig struct {
	Name string            `yaml:"name"`
	Env  map[string]string `yaml:"env"`
}

func CreateEnvironmentConfig(basePath string) error {
	path := environmentConfigPath(basePath)
	if err := ioutil.WriteFile(path, []byte(environmentConfigExample), 0644); err != nil {
		return log.Wrap(err)
	}
	return nil
}

const environmentConfigExample = `# Here you can define environment variables for functions
# on a project, stage or function level. If the same variable is
# defined on multiple levels the lowest level will take precedence.
# For example, uncommenting the config below will result in
# the function ping having the following environment:
# KEY: project
# KEY2: stage
# KEY3: function

# project:
#   env:
#     KEY: project
#     KEY2: project
#     KEY3: project
#   stages: 
#     - name: dev
#       env:
#         KEY2: stage
#         KEY3: stage
#       functions:
#       - name: ping
#         env:
#           KEY3: function
`

func environmentConfigPath(basePath string) string {
	return filepath.Join(basePath, configDir, environmentConfigName)
}
