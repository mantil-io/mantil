package workspace

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/cli/log"
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
	Name        string   `yaml:"name"`
	Stages      []*Stage `yaml:"stages,omitempty"`
	workspace   *Workspace
	environment *EnvironmentConfig
}

func (p *Project) ResourceTags() map[string]string {
	return map[string]string{
		TagProjectName: p.Name,
	}
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

func (p *Project) NewStage(stageName, accountName string) (*Stage, error) {

	if stageName == "" {
		stageName = DefaultStageName
	}
	if p.Stage(stageName) != nil {
		return nil, ErrStageExists
	}
	account := p.workspace.FindAccount(accountName)
	if account == nil {
		return nil, ErrAccountNotFound
	}
	stage := &Stage{
		Name:        stageName,
		AccountName: account.Name,
		Public:      &Public{},
		account:     account,
		project:     p,
	}
	if len(p.Stages) == 0 {
		stage.Default = true
	}
	p.Stages = append(p.Stages, stage)
	return stage, nil
}

func (p *Project) RemoveStage(stageName string) {
	for idx, s := range p.Stages {
		if s.Name == stageName {
			p.Stages = append(p.Stages[:idx], p.Stages[idx+1:]...)
		}
	}
}

// TODO remove use function.LambdaName
func ProjectResource(projectName string, v ...string) string {
	r := projectName
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}

func (p *Project) LogGroupPrefix() string {
	return p.Name
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
