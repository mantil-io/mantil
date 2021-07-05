package main

import (
	"fmt"
	"strings"

	"github.com/atoz-technology/mantil-cli/pkg/shell"
)

type Spa struct {
	Organization Organization // server side only

	Name string // required
	//ApiDomain string // optional, default = api
	//AppDomain string // optional, default = www

	//	Users     []User
	Functions []Function
}

// type User struct {
// 	Username string
// 	Email    string
// }

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
	// if s.ApiDomain == "" {
	// 	s.ApiDomain = "api"
	// }
	// if s.AppDomain == "" {
	// 	s.AppDomain = "www"
	// }
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
		f.URL = fmt.Sprintf("https://%s/%s", p.ApiURL(), f.Path)
		p.Functions[i] = f
	}
}

func (s Spa) ApiURL() string {
	return fmt.Sprintf("%s.%s", s.Name, s.Organization.DNSZone)
}

// func (s Spa) AppURL() string {
// 	return fmt.Sprintf("%s.%s.%s", s.AppDomain, s.Name, s.Organization.DNSZone)
// }

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

		// Users: []User{
		// 	{Username: "ianic", Email: "ianic@atoz.technology"},
		// 	{Username: "patko", Email: "patko@atoz.technology"},
		// },
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

	// if err := s.findCertificate(); err != nil {
	// 	return err
	// }
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

// func (s *Spa) findCertificate() error {
// 	// example of how to find certificate for project
// 	var cert Cert
// 	if err := s.Organization.LoadProject("cert", &cert); err != nil {
// 		// TODO if cert project dont exists create one
// 		return err
// 	}
// 	arn := cert.Arn(s.Name)
// 	if arn == "" {
// 		// TODO create certificate
// 		return fmt.Errorf("certificate for project %s not found", s.Name)
// 	}

// 	log.Printf("cert arn found: %s", arn)
// 	s.CertArn = arn

// 	return nil
// }

// func (s *Spa) Get() (*Spa, error) {
// 	return s, nil
// }
