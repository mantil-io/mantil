package mantil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mantil-io/mantil-backend/internal/aws"
)

const (
	configS3Key     = "config/project.json"
	localConfigPath = "config/mantil.local.json"
	defaultStage    = "dev"
)

const (
	envProjectName = "MANTIL_PROJECT_NAME"
	envStageName   = "MANTIL_STAGE_NAME"
)

type Project struct {
	Organization   Organization
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
	ImageKey   string
	Runtime    string
	Handler    string
	MemorySize int
	Timeout    int
	Env        map[string]string
	Path       string
	URL        string
	Public     bool
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
	Name     string
	Hash     string
	S3Key    string
	ImageKey string
}

type StaticWebsiteUpdate struct {
	Name string
	Hash string
}

func TryOrganization() Organization {
	return Organization{
		Name:    "try",
		DNSZone: "try.mantil.team",
		CertArn: "arn:aws:acm:us-east-1:477361877445:certificate/f412a03f-ad0f-473c-b4ba-0b513b423c36",
	}
}

func ProjectResourceId(projectName string) string {
	org := TryOrganization()
	return fmt.Sprintf("%s-%s", strings.Replace(org.DNSZone, ".", "-", -1), projectName)
}

func ProjectIdentifier(projectName string) string {
	org := TryOrganization()
	return fmt.Sprintf("mantil-project-%s-%s", org.Name, projectName)
}

func ProjectBucket(projectName string) string {
	return ProjectIdentifier(projectName)
}

func NewProject(name, token string) (*Project, error) {
	org := TryOrganization()
	p := &Project{
		Organization: org,
		Name:         name,
		Bucket:       ProjectBucket(name),
		Token:        token,
	}
	return p, nil
}

func LoadProject(projectName string) (*Project, error) {
	bucket := ProjectBucket(projectName)
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
		if f.S3Key == "" && f.ImageKey == "" {
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
		f.URL = fmt.Sprintf("https://%s/%s/%s", p.Organization.DNSZone, p.Name, f.Path)
		if f.Env == nil {
			f.Env = make(map[string]string)
		}
		f.Env[envProjectName] = p.Name
		f.Env[envStageName] = defaultStage
		p.Functions[i] = f
	}
}

func (p *Project) IsValidToken(token string) bool {
	return p.Token == token
}
