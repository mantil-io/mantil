package mantil

import (
	"fmt"
	"strings"

	"github.com/atoz-technology/mantil-cli/pkg/shell"
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

func NewProject(name string) Project {
	org := TryOrganization()
	p := Project{
		Organization: org,
		Name:         name,
		Bucket:       fmt.Sprintf("mantil-project-%s-%s", org.Name, name),
	}
	return p
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
	p.addDefaults()
}

func (p *Project) Dummy() (interface{}, error) {
	return nil, nil
}

func (p *Project) Test() (interface{}, error) {
	p.TestData()
	return nil, p.Apply()
}

func (p *Project) Apply() error {
	org := p.Organization
	if err := org.PrepareProject("go-func", p.Name, p); err != nil {
		return err
	}
	tf := shell.Terraform(org.ProjectFolder(p.Name))
	if err := tf.Init(); err != nil {
		return err
	}
	if err := tf.Plan(); err != nil {
		return err
	}
	if err := tf.Apply(); err != nil {
		return err
	}
	if err := org.PushProject(p.Name, p); err != nil {
		return err
	}

	return nil
}
