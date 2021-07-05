package main

import (
	"fmt"
	"strings"

	"github.com/atoz-technology/mantil-cli/pkg/shell"
)

type Spa struct {
	Organization Organization
	Name         string // required
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

func (s *Spa) addDefaults() {
	s.addFunctionDefaults()
}

func (p *Spa) addFunctionDefaults() {
	for i, f := range p.Functions {
		if f.Path == "" {
			f.Path = f.Name
		}
		if f.S3Key == "" {
			f.S3Key = fmt.Sprintf("%s.zip", f.Name)
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
			if strings.HasPrefix(f.Runtime, "node") {
				f.Handler = "index.handler"
			}
			if strings.HasPrefix(f.Runtime, "go") {
				f.Handler = "main"
			}
		}
		f.URL = fmt.Sprintf("https://%s/%s/%s", p.Organization.DNSZone, p.Name, f.Path)
		p.Functions[i] = f
	}
}

func tryOrganization() Organization {
	return Organization{
		Name:            "try",
		DNSZone:         "try.mantil.team",
		FunctionsBucket: "try.mantil.team-lambda-functions",
		CertArn:         "arn:aws:acm:us-east-1:477361877445:certificate/f412a03f-ad0f-473c-b4ba-0b513b423c36",
	}
}

func (s *Spa) testData() {
	*s = Spa{
		Organization: tryOrganization(),
		Name:         "proj1",
		Functions: []Function{
			{
				Name:    "hello",
				Runtime: "go1.x",
				S3Key:   "functions/hello:v016b704-dirty.zip",
				Timeout: 60,
				Public:  true,
				//Env:     map[string]string{"foo": "bar"},
			},
			// {
			// 	Name:    "second",
			// 	S3Key:   "functions/second:v0.1.1.zip",
			// 	Runtime: "nodejs14.x",
			// 	Timeout: 60,
			// },
		},
	}
	s.addDefaults()
}

func (s *Spa) Dummy() (interface{}, error) {
	return nil, nil
}

func (s *Spa) Test() (interface{}, error) {
	s.testData()
	return nil, s.Apply()
}

func (s *Spa) Apply() error {
	org := s.Organization
	if err := org.PrepareProject("go-func", s.Name, s); err != nil {
		return err
	}
	tf := shell.Terraform(org.ProjectFolder(s.Name))
	if err := tf.Init(); err != nil {
		return err
	}
	if err := tf.Plan(); err != nil {
		return err
	}
	if err := tf.Apply(); err != nil {
		return err
	}
	if err := org.PushProject(s.Name, s); err != nil {
		return err
	}

	return nil
}
