package workspace

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/mantil-io/mantil/cli/log"
	"gopkg.in/yaml.v2"
)

const (
	DefaultStageName = "dev"
	// env variable containing comma separated list of env variables
	// that should be added as tags to all resource created by stage lambda functions
	EnvMantilStageTags = "MANTIL_ENV_TAGS"
	TagStageName       = "MANTIL_STAGE"
)

var (
	// list of env variables for EnvMantilStageTags
	// variables specified here should be passed to all lambda functions
	MantilStageTags = []string{EnvProjectName, EnvStageName}
)

type Stage struct {
	Name      string          `yaml:"name"`
	Default   bool            `yaml:"default,omitempty"`
	Account   string          `yaml:"account"`
	Endpoints *StageEndpoints `yaml:"endpoints,omitempty"`
	Functions []*Function     `yaml:"functions,omitempty"`
	Public    *Public         `yaml:"public,omitempty"`
}

type Public struct {
	Bucket string        `yaml:"bucket"`
	Sites  []*PublicSite `yaml:"sites"`
}

type PublicSite struct {
	Name string `yaml:"name"`
	Hash string `yaml:"hash"`
}

func (s *Stage) ResourceTags() map[string]string {
	return map[string]string{
		TagStageName: s.Name,
	}
}

type StageEndpoints struct {
	Rest string `yaml:"rest"`
	Ws   string `yaml:"ws"`
}

func StageBucketPrefix(projectName, stageName string) string {
	return fmt.Sprintf("stages/%s/%s", projectName, stageName)
}

func (s *Stage) ApplyEnv(basePath, projectName, workspaceKey string) (bool, error) {
	path := environmentConfigPath(basePath)
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return false, log.Wrap(err)
	}
	ec := &EnvironmentConfig{}
	if err := yaml.Unmarshal(buf, ec); err != nil {
		return false, log.Wrap(err)
	}
	changed := false
	for _, f := range s.Functions {
		envChain := []map[string]string{
			s.defaultEnv(projectName, workspaceKey),
			ec.Project.Env,
		}
		for _, sc := range ec.Project.Stages {
			if sc.Name != s.Name {
				continue
			}
			envChain = append(envChain, sc.Env)
			for _, fc := range sc.Functions {
				if f.Name != fc.Name {
					continue
				}
				envChain = append(envChain, fc.Env)
			}
		}
		// reverse the chain to get correct priorities
		for i, j := 0, len(envChain)-1; i < j; i, j = i+1, j-1 {
			envChain[i], envChain[j] = envChain[j], envChain[i]
		}
		changed = f.mergeEnv(envChain...)
	}
	return changed, nil
}

func (s *Stage) defaultEnv(projectName, workspaceKey string) map[string]string {
	env := map[string]string{
		EnvProjectName:     projectName,
		EnvStageName:       s.Name,
		EnvWorkspaceKey:    workspaceKey,
		EnvMantilStageTags: strings.Join(MantilStageTags, ","),
	}
	return env
}
