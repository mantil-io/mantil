package mantil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/mantil-io/mantil/internal/aws"
)

const (
	configS3Key     = "config/project.json"
	localConfigPath = "config/mantil.local.json"
	defaultStage    = "dev"
)

const (
	EnvProjectName = "MANTIL_PROJECT_NAME"
	EnvStageName   = "MANTIL_STAGE_NAME"
	EnvApiURL      = "MANTIL_API_URL"
)

type Project struct {
	Name           string // required
	Bucket         string
	Token          string
	ApiURL         string
	Functions      []Function
	StaticWebsites []StaticWebsite
}

type Function struct {
	Name       string
	Hash       string
	S3Key      string
	Runtime    string
	Handler    string
	MemorySize int
	Timeout    int
	Path       string
	Env        map[string]string
}

type StaticWebsite struct {
	Name   string
	Bucket string
	Hash   string
}

type UpdateAction uint8

const (
	Add    UpdateAction = 0
	Remove UpdateAction = 1
	Update UpdateAction = 2
)

type ProjectUpdate struct {
	Function      *FunctionUpdate
	StaticWebsite *StaticWebsiteUpdate
	Action        UpdateAction
}

type FunctionUpdate struct {
	Name  string
	Hash  string
	S3Key string
}

type StaticWebsiteUpdate struct {
	Name string
	Hash string
}

func (f *Function) SetS3Key(key string) {
	f.S3Key = key
}

func ProjectBucket(projectName string, aws *aws.AWS) (string, error) {
	accountID, err := aws.AccountID()
	if err != nil {
		return "", err
	}
	return ProjectResource(projectName, accountID), nil
}

func NewProject(name, token string, aws *aws.AWS) (*Project, error) {
	bucket, err := ProjectBucket(name, aws)
	if err != nil {
		return nil, err
	}
	p := &Project{
		Name:   name,
		Bucket: bucket,
		Token:  token,
	}
	return p, nil
}

func ProjectResource(projectName string, v ...string) string {
	r := fmt.Sprintf("mantil-project-%s", projectName)
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
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

type LocalProjectConfig struct {
	Name   string `json:"name"`
	ApiURL string `json:"apiURL,omitempty"`
}

func LocalConfig(name string) *LocalProjectConfig {
	return &LocalProjectConfig{
		Name: name,
	}
}

func (c *LocalProjectConfig) Save(path string) error {
	buf, err := json.Marshal(c)
	if err != nil {
		return err
	}
	configDir := filepath.Join(path, "config")
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(path, localConfigPath), buf, 0644); err != nil {
		return err
	}
	return nil
}

func LoadLocalConfig(projectRoot string) (*LocalProjectConfig, error) {
	buf, err := ioutil.ReadFile(filepath.Join(projectRoot, localConfigPath))
	if err != nil {
		return nil, err
	}
	c := &LocalProjectConfig{}
	if err := json.Unmarshal(buf, c); err != nil {
		return nil, err
	}
	return c, nil
}

func LoadProject(projectName string) (*Project, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	bucket, err := ProjectBucket(projectName, awsClient)
	if err != nil {
		return nil, err
	}
	p := &Project{}
	if err := awsClient.GetObjectFromS3Bucket(bucket, configS3Key, p); err != nil {
		return nil, err
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
	if err := awsClient.PutObjectToS3Bucket(p.Bucket, configS3Key, buf); err != nil {
		return err
	}
	return nil
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

func Env() (string, *LocalProjectConfig) {
	initPath := "."
	path, err := FindProjectRoot(initPath)
	if err != nil {
		log.Fatal(err)
	}
	config, err := LoadLocalConfig(path)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf(`export %s='%s'
export %s='%s'
`, EnvProjectName, config.Name,
		EnvApiURL, config.ApiURL,
	), config
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
