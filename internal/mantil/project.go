package mantil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/atoz-technology/mantil-cli/internal/aws"
)

const (
	configS3Key     = "config/project.json"
	localConfigPath = "config/mantil.local.json"
)

type Project struct {
	Organization Organization
	Name         string // required
	Bucket       string
	AccessTag    string
	Functions    []Function
}

type Function struct {
	Name       string
	Hash       string
	S3Key      string
	Runtime    string
	Handler    string
	MemorySize int
	Timeout    int
	Env        map[string]string
	Path       string
	URL        string
	Public     bool
}

func TryOrganization() Organization {
	return Organization{
		Name:    "try",
		DNSZone: "try.mantil.team",
		CertArn: "arn:aws:acm:us-east-1:477361877445:certificate/f412a03f-ad0f-473c-b4ba-0b513b423c36",
	}
}

func ProjectBucket(projectName string) string {
	org := TryOrganization()
	return fmt.Sprintf("mantil-project-%s-%s", org.Name, projectName)
}

func AccessTag(projectName string) string {
	org := TryOrganization()
	return fmt.Sprintf("%s-%s", org.Name, projectName)
}

func NewProject(name string) (*Project, error) {
	org := TryOrganization()
	p := &Project{
		Organization: org,
		Name:         name,
		Bucket:       ProjectBucket(name),
		AccessTag:    AccessTag(name),
	}
	return p, nil
}

func LoadProject(bucket string) (*Project, error) {
	awsClient, err := aws.New()
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
	if err := awsClient.PutObjectToS3Bucket(p.Bucket, configS3Key, bytes.NewReader(buf)); err != nil {
		return err
	}
	return nil
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
			f.Runtime = "go1.x"
		}
		if f.MemorySize == 0 {
			f.MemorySize = 128
		}
		if f.Timeout == 0 {
			f.Timeout = 60 * 15
		}
		if f.Handler == "" {
			f.Handler = f.Name
		}
		f.URL = fmt.Sprintf("https://%s/%s/%s", p.Organization.DNSZone, p.Name, f.Path)
		p.Functions[i] = f
	}
}

type LocalProjectConfig struct {
	Bucket string
}

func (p *Project) LocalConfig() *LocalProjectConfig {
	return &LocalProjectConfig{
		Bucket: p.Bucket,
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
