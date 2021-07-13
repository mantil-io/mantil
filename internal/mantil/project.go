package mantil

import (
	"fmt"
	"io/ioutil"
	"strings"
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

func NewProject(name, funcsPath string) (*Project, error) {
	org := TryOrganization()
	p := &Project{
		Organization: org,
		Name:         name,
		Bucket:       fmt.Sprintf("mantil-project-%s-%s", org.Name, name),
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

func (p *Project) addDefaults() {
	p.addFunctionDefaults()
}

func (p *Project) addFunctionDefaults() {
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

func TryOrganization() Organization {
	return Organization{
		Name:    "try",
		DNSZone: "try.mantil.team",
		CertArn: "arn:aws:acm:us-east-1:477361877445:certificate/f412a03f-ad0f-473c-b4ba-0b513b423c36",
	}
}

func (p *Project) TestData() {
	*p = Project{
		Organization: TryOrganization(),
		Name:         "proj1",
		Functions: []Function{
			{
				Name:    "hello",
				Runtime: "go1.x",
				S3Key:   "functions/hello:v016b704-dirty.zip",
				Timeout: 60,
				Public:  true,
			},
		},
	}
	p.addDefaults()
}
