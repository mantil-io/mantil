package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/aws"
	"gopkg.in/yaml.v2"
)

const (
	configDir    = "config"
	configName   = "project.yml"
	configS3Key  = "config/project.yml"
	defaultStage = "dev"
	TokenLength  = 40
)

const (
	EnvProjectName = "MANTIL_PROJECT_NAME"
	EnvStageName   = "MANTIL_STAGE_NAME"
	EnvApiURL      = "MANTIL_API_URL"
)

type Project struct {
	Name        string        `yaml:"name"` // required
	Bucket      string        `yaml:"bucket"`
	Functions   []*Function   `yaml:"functions"`
	PublicSites []*PublicSite `yaml:"public_sites"`
	Stages      []*Stage      `yaml:"stages"`
}

type ProjectUpdate struct {
	Function   *FunctionUpdate
	PublicSite *PublicSiteUpdate
	Action     UpdateAction
}

type UpdateAction uint8

const (
	Add    UpdateAction = 0
	Remove UpdateAction = 1
	Update UpdateAction = 2
)

func CreateProject(name string, aws *aws.AWS) (*Project, error) {
	project, err := NewProject(name, aws)
	if err != nil {
		return nil, err
	}
	if err := SaveProjectS3(project); err != nil {
		return nil, fmt.Errorf("could not save project configuration - %v", err)
	}
	return project, nil
}

func NewProject(name string, aws *aws.AWS) (*Project, error) {
	bucket, err := Bucket(aws)
	if err != nil {
		return nil, err
	}
	p := &Project{
		Name:   name,
		Bucket: bucket,
	}
	return p, nil
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

func SaveProjectS3(p *Project) error {
	awsClient, err := aws.New()
	if err != nil {
		return err
	}
	buf, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	if err := awsClient.PutObjectToS3Bucket(p.Bucket, ProjectS3ConfigKey(p.Name), buf); err != nil {
		return err
	}
	return nil
}

func LoadProjectS3(projectName string) (*Project, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	bucket, err := Bucket(awsClient)
	if err != nil {
		return nil, err
	}
	buf, err := awsClient.GetObjectFromS3Bucket(bucket, ProjectS3ConfigKey(projectName))
	if err != nil {
		return nil, err
	}
	p := &Project{}
	if err := yaml.Unmarshal(buf, p); err != nil {
		return nil, err
	}
	return p, nil
}

func ProjectS3ConfigKey(projectName string) string {
	return fmt.Sprintf("%s%s", ProjectBucketPrefix(projectName), configS3Key)
}

func DeleteProjectStage(p *Project, stage string, aws *aws.AWS) error {
	return aws.DeleteInS3Bucket(p.Bucket, p.StageBucketPrefix(stage))
}

func DeleteProject(p *Project, aws *aws.AWS) error {
	return aws.DeleteInS3Bucket(p.Bucket, ProjectBucketPrefix(p.Name))
}

func configPath(basePath string) string {
	return filepath.Join(basePath, configDir, configName)
}

func ProjectBucketPrefix(projectName string) string {
	return fmt.Sprintf("projects/%s/", projectName)
}

func (p *Project) StageBucketPrefix(stage string) string {
	return fmt.Sprintf("%s%s/", ProjectBucketPrefix(p.Name), stage)
}

func (p *Project) Stage(name string) *Stage {
	for _, s := range p.Stages {
		if s.Name == name {
			return s
		}
	}
	return nil
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

func (p *Project) RestEndpoint(stageName string) string {
	s := p.Stage(stageName)
	if s == nil || s.Endpoints == nil {
		return ""
	}
	return s.Endpoints.Rest
}

func ProjectResource(projectName string, stageName string, v ...string) string {
	r := fmt.Sprintf("mantil-project-%s-%s", projectName, stageName)
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}

func ProjectExists(name string, aws *aws.AWS) (bool, error) {
	bucket, err := Bucket(aws)
	if err != nil {
		return false, err
	}
	return aws.S3PrefixExistsInBucket(bucket, ProjectBucketPrefix(name))
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

func (p *Project) AddFunction(fun *Function) {
	p.Functions = append(p.Functions, fun)
}

func (p *Project) RemoveFunction(fun string) {
	for i, f := range p.Functions {
		if fun == f.Name {
			p.Functions = append(p.Functions[:i], p.Functions[i+1:]...)
			break
		}
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
