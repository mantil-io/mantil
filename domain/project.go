package domain

import (
	"fmt"

	"github.com/mantil-io/mantil/kit/schema"
	"github.com/mantil-io/mantil/kit/token"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	StateBucketPrefix         = "state"
	FunctionsBucketPrefix     = "functions"
	FunctionsBucketExpireDays = 7
)

type Project struct {
	Name        string   `yaml:"name"`
	Stages      []*Stage `yaml:"stages,omitempty"`
	workspace   *Workspace
	environment *EnvironmentConfig
}

func (p *Project) resourceTags() map[string]string {
	return map[string]string{
		TagProjectName: p.Name,
	}
}

func (p *Project) Stage(name string) *Stage {
	for _, s := range p.Stages {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (p *Project) DefaultStage() *Stage {
	for _, s := range p.Stages {
		if s.Default {
			return s
		}
	}
	return nil
}

func (p *Project) SetDefaultStage(name string) {
	if len(p.Stages) == 0 {
		return
	}
	s := p.Stage(name)
	if s == nil || s.Default {
		return
	}
	for _, ps := range p.Stages {
		if ps.Default {
			ps.Default = false
		}
	}
	s.Default = true
}

func (p *Project) setDefaultStage() {
	if len(p.Stages) == 0 {
		return
	}
	if s := p.DefaultStage(); s != nil {
		return
	}
	p.Stages[0].Default = true
}

func (p *Project) FunctionsBucketPrefix() string {
	return fmt.Sprintf("%s/%s", FunctionsBucketPrefix, p.Name)
}

func (p *Project) StateBucketPrefix() string {
	return fmt.Sprintf("%s/%s", StateBucketPrefix, p.Name)
}

func (p *Project) NewStage(stageName, nodeName, path string) (*Stage, error) {
	if stageName == "" {
		stageName = DefaultStageName
	}
	if err := ValidateName(stageName); err != nil {
		return nil, err
	}
	if p.Stage(stageName) != nil {
		return nil, &StageExistsError{stageName}
	}
	node := p.workspace.FindNode(nodeName)
	if node == nil {
		return nil, &NodeNotFoundError{nodeName}
	}
	publicKey, privateKey, err := token.KeyPair()
	if err != nil {
		return nil, errors.Wrap(err, "could not create public/private key pair")
	}
	stage := &Stage{
		Name:     stageName,
		NodeName: node.Name,
		Keys: StageKeys{
			Public:  publicKey,
			Private: privateKey,
		},
		node:    node,
		project: p,
	}
	if len(p.Stages) == 0 {
		stage.Default = true
	}
	p.Stages = append(p.Stages, stage)
	stage.node.AddStage(stage.Name, p.Name, path)
	if p.NumberOfStages() == 1 {
		stage.node.workspace.AddProject(p.Name, path)
	}
	return stage, nil
}

func (p *Project) RemoveStage(stageName string) {
	for idx, s := range p.Stages {
		if s.Name == stageName {
			p.Stages = append(p.Stages[:idx], p.Stages[idx+1:]...)
			s.node.RemoveStage(s.Name)
			if p.NumberOfStages() == 0 {
				s.node.workspace.RemoveProject(p.Name)
			}
		}
	}
	p.setDefaultStage()
}

func (p *Project) NumberOfStages() int {
	return len(p.Stages)
}

func (p *Project) numberOfNodes() int {
	m := make(map[string]bool)
	for _, s := range p.Stages {
		m[s.NodeName] = true
	}
	return len(m)
}

func (p *Project) numberOfAWSAccountsAndRgions() (int, int) {
	a := make(map[string]bool)
	r := make(map[string]bool)
	for _, s := range p.Stages {
		a[s.node.AccountID] = true
		r[s.node.Region] = true
	}
	return len(a), len(r)
}

func (p *Project) numberOfFunctions() int {
	m := make(map[string]bool)
	for _, s := range p.Stages {
		for _, f := range s.Functions {
			m[f.Name] = true
		}
	}
	return len(m)
}

func (p *Project) AsCliProject() *CliProject {
	if p == nil {
		return nil
	}
	a, r := p.numberOfAWSAccountsAndRgions()
	return &CliProject{
		Name:        p.Name,
		Stages:      p.NumberOfStages(),
		Nodes:       p.numberOfNodes(),
		AWSAccounts: a,
		AWSRegions:  r,
	}
}

type EnvironmentConfig struct {
	Project ProjectEnvironmentConfig `yaml:"project" jsonschema:"nullable,default={}"`
}

type ProjectEnvironmentConfig struct {
	Stages                []StageEnvironmentConfig `yaml:"stages,omitempty" jsonschema:"nullable,default=[]"`
	FunctionConfiguration `yaml:",inline"`
}

func (c ProjectEnvironmentConfig) StageEnvConfig(name string) StageEnvironmentConfig {
	for _, s := range c.Stages {
		if s.Name == name {
			return s
		}
	}
	return StageEnvironmentConfig{}
}

type StageEnvironmentConfig struct {
	Name                  string                      `yaml:"name"`
	Functions             []FunctionEnvironmentConfig `yaml:"functions,omitempty"`
	FunctionConfiguration `yaml:",inline"`
	CustomDomain          CustomDomain `yaml:"custom_domain,omitempty" jsonschema:"nullable,default={}"`
}

type CustomDomain struct {
	DomainName       string `yaml:"domain_name"`
	CertDomain       string `yaml:"cert_domain,omitempty"`
	HostedZoneDomain string `yaml:"hosted_zone_domain,omitempty"`
	HttpSubdomain    string `yaml:"http_subdomain,omitempty"`
	WsSubdomain      string `yaml:"ws_subdomain,omitempty"`
}

func (cd *CustomDomain) setDefaults() {
	if cd.DomainName == "" {
		return
	}
	if cd.CertDomain == "" {
		cd.CertDomain = cd.DomainName
	}
	if cd.HostedZoneDomain == "" {
		cd.HostedZoneDomain = cd.DomainName
	}
	if cd.HttpSubdomain == "" {
		cd.HttpSubdomain = "api"
	}
	if cd.WsSubdomain == "" {
		cd.WsSubdomain = "ws"
	}
}

func (c StageEnvironmentConfig) FunctionEnvConfig(name string) FunctionEnvironmentConfig {
	for _, f := range c.Functions {
		if f.Name == name {
			return f
		}
	}
	return FunctionEnvironmentConfig{}
}

type FunctionEnvironmentConfig struct {
	Name                  string `yaml:"name"`
	FunctionConfiguration `yaml:",inline"`
}

const environmentConfigExample = `# Here you can define various configuration parameters for functions
# such as environment variables, memory size, execution schedule and timeout duration.
# These can be defined on a project, stage or function level. If the same parameter is
# defined on multiple levels the lowest level will take precedence.
# For example, uncommenting the config below will result in
# the function ping having the following configuration:
# memory_size: 512
# timeout: 60
# cron: "* * * * ? *"
# env:
#   KEY: project
#   KEY2: stage
#   KEY3: function

# project:
#   memory_size: 128
#   timeout: 30
#   env:
#     KEY: project
#     KEY2: project
#     KEY3: project
#   stages: 
#     - name: dev
#       memory_size: 256
#       timeout: 60
#       env:
#         KEY2: stage
#         KEY3: stage
#       functions:
#       - name: ping
#         memory_size: 512
#         cron: "* * * * ? *"
#         private: true
#         env:
#           KEY3: function
`

func ValidateEnvironmentConfig(buf []byte) (*EnvironmentConfig, error) {
	ec := &EnvironmentConfig{}
	schema, err := schema.From(ec)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := schema.ValidateYAML(buf); err != nil {
		return nil, &EnvironmentConfigValidationError{err}
	}
	if err := yaml.Unmarshal(buf, ec); err != nil {
		return nil, errors.WithStack(err)
	}
	if !ec.validateCron() {
		return nil, &EnvironmentConfigValidationError{
			fmt.Errorf("invalid cron syntax"),
		}
	}
	return ec, nil
}

func (ec *EnvironmentConfig) validateCron() bool {
	p := ec.Project
	if !p.validateCron() {
		return false
	}
	for _, s := range p.Stages {
		if !s.validateCron() {
			return false
		}
		for _, f := range s.Functions {
			if !f.validateCron() {
				return false
			}
		}
	}
	return true
}
