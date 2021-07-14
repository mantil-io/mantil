package mantil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/atoz-technology/mantil-cli/internal/aws"
)

type Project struct {
	Organization Organization
	Name         string // required
	Bucket       string
	Functions    []Function
}

type Function struct {
	Name       string
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

func NewProject(name, funcsPath string) (*Project, error) {
	org := TryOrganization()
	p := &Project{
		Organization: org,
		Name:         name,
		Bucket:       ProjectBucket(name),
	}
	if funcsPath == "" {
		return p, nil
	}
	files, err := ioutil.ReadDir(funcsPath)
	if err != nil {
		return nil, err
	}
	// go through functions in functions directory
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		name := f.Name()
		f := Function{
			Path:       name,
			Name:       name,
			S3Key:      fmt.Sprintf("functions/%s.zip", name),
			Runtime:    "go1.x",
			MemorySize: 128,
			Timeout:    60 * 15,
			Handler:    name,
		}
		f.URL = fmt.Sprintf("https://%s/%s/%s", p.Organization.DNSZone, p.Name, f.Path)
		p.Functions = append(p.Functions, f)
	}
	return p, nil
}

const configS3Key = "config/project.json"

func LoadProject(name string) (*Project, error) {
	bucket := ProjectBucket(name)
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
