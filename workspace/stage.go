package workspace

import (
	"fmt"
	"strings"

	"github.com/mantil-io/mantil/aws"
	"gopkg.in/yaml.v2"
)

const (
	DefaultStageName = "dev"
	// env variable containing comma separated list of env variables
	// that should be added as tags to all resource created by stage lambda functions
	EnvMantilStageTags = "MANTIL_ENV_TAGS"
	TagStageName       = "MANTIL_STAGE"
	TagStageKey        = "MANTIL_KEY"
)

var (
	// list of env variables for EnvMantilStageTags
	// variables specified here should be passed to all lambda functions
	MantilStageTags = []string{EnvProjectName, EnvStageName}
)

type Stage struct {
	Name      string            `yaml:"name"`
	Default   bool              `yaml:"default,omitempty"`
	Account   string            `yaml:"account"`
	Endpoints *StageEndpoints   `yaml:"endpoints,omitempty"`
	Env       map[string]string `yaml:"env,omitempty"`
	Functions []*Function       `yaml:"functions,omitempty"`
	Public    *Public           `yaml:"public,omitempty"`
}

func (s *Stage) ResourceTags() map[string]string {
	return map[string]string{
		TagStageName: s.Name,
		TagStageKey:  "key-placeholder",
	}
}

type StageEndpoints struct {
	Rest string `yaml:"rest"`
	Ws   string `yaml:"ws"`
}

// TODO: remove when you remove data function
func LoadStageState(bucket, projectName, stageName string) (*Stage, error) {
	aws, err := aws.New()
	if err != nil {
		return nil, err
	}
	s3Key := StageStateS3Key(projectName, stageName)
	buf, err := aws.GetObjectFromS3Bucket(bucket, s3Key)
	if err != nil {
		return nil, err
	}
	s := &Stage{}
	if err := yaml.Unmarshal(buf, s); err != nil {
		return nil, err
	}
	return s, nil
}

// TODO: remove when you remove data function
func StageStateS3Key(projectName, stageName string) string {
	return fmt.Sprintf("%s/state.yml", StageBucketPrefix(projectName, stageName))
}

func StageEnv(projectName, stageName string) map[string]string {
	env := map[string]string{
		EnvProjectName:     projectName,
		EnvStageName:       stageName,
		EnvMantilStageTags: strings.Join(MantilStageTags, ","),
	}
	return env
}

func StageBucketPrefix(projectName, stageName string) string {
	return fmt.Sprintf("stages/%s/%s", projectName, stageName)
}
