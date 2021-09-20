package mantil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/util"
)

const (
	configS3Key             = "config/project.json"
	projectS3PrefixTemplate = "projects/%s/"
	defaultStage            = "dev"
	TokenLength             = 40
)

const (
	EnvProjectName = "MANTIL_PROJECT_NAME"
	EnvStageName   = "MANTIL_STAGE_NAME"
	EnvApiURL      = "MANTIL_API_URL"
)

type Project struct {
	Name           string // required
	Bucket         string
	BucketPrefix   string
	Token          string
	ApiURL         string
	Functions      []Function
	StaticWebsites []StaticWebsite
}

type ProjectUpdate struct {
	Function      *FunctionUpdate
	StaticWebsite *StaticWebsiteUpdate
	Action        UpdateAction
}

type UpdateAction uint8

const (
	Add    UpdateAction = 0
	Remove UpdateAction = 1
	Update UpdateAction = 2
)

func CreateProject(name string, aws *aws.AWS) (*Project, error) {
	token := util.GenerateToken(TokenLength)
	if token == "" {
		return nil, fmt.Errorf("could not generate token for project %s", name)
	}
	project, err := NewProject(name, token, aws)
	if err != nil {
		return nil, err
	}
	if err := SaveProject(project); err != nil {
		return nil, fmt.Errorf("could not save project configuration - %v", err)
	}
	return project, nil
}

func NewProject(name, token string, aws *aws.AWS) (*Project, error) {
	bucket, err := Bucket(aws)
	if err != nil {
		return nil, err
	}
	p := &Project{
		Name:         name,
		Bucket:       bucket,
		BucketPrefix: ProjectBucketPrefix(name),
		Token:        token,
	}
	return p, nil
}

func SaveProject(p *Project) error {
	awsClient, err := aws.New()
	if err != nil {
		return err
	}
	buf, err := json.Marshal(p)
	if err != nil {
		return err
	}
	if err := awsClient.PutObjectToS3Bucket(p.Bucket, ProjectS3ConfigKey(p.Name), buf); err != nil {
		return err
	}
	return nil
}

func LoadProject(projectName string) (*Project, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	bucket, err := Bucket(awsClient)
	if err != nil {
		return nil, err
	}
	p := &Project{}
	if err := awsClient.GetObjectFromS3Bucket(bucket, ProjectS3ConfigKey(projectName), p); err != nil {
		return nil, err
	}
	return p, nil
}

func DeleteProject(p *Project, aws *aws.AWS) error {
	return aws.DeleteInS3Bucket(p.Bucket, p.BucketPrefix)
}

func ProjectBucketPrefix(projectName string) string {
	return fmt.Sprintf(projectS3PrefixTemplate, projectName)
}

func ProjectResource(projectName string, v ...string) string {
	r := fmt.Sprintf("mantil-project-%s", projectName)
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}

func ProjectS3ConfigKey(projectName string) string {
	return fmt.Sprintf("%s%s", ProjectBucketPrefix(projectName), configS3Key)
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
		_, err := os.Stat(filepath.Join(currentPath, localConfigPath))
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

func (p *Project) AddFunction(fun Function) {
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

func (p *Project) S3FileKey(file string) string {
	return fmt.Sprintf("%s%s", p.BucketPrefix, file)
}

func (p *Project) AddFunctionDefaults() {
	for i, f := range p.Functions {
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
		f.Env[EnvProjectName] = p.Name
		f.Env[EnvStageName] = defaultStage
		p.Functions[i] = f
	}
}

func (p *Project) IsValidToken(token string) bool {
	return p.Token == token
}
