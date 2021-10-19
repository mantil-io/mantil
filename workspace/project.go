package workspace

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

func (p *Project) setDefaultStage() {
	if len(p.Stages) == 0 {
		return
	}
	if s := p.DefaultStage(); s != nil {
		return
	}
	p.Stages[0].Default = true
}

func (p *Project) NewStage(stageName, accountName string) (*Stage, error) {
	if stageName == "" {
		stageName = defaultStageName
	}
	if p.Stage(stageName) != nil {
		return nil, ErrStageExists
	}
	account := p.workspace.FindAccount(accountName)
	if account == nil {
		return nil, ErrAccountNotFound
	}
	stage := &Stage{
		Name:        stageName,
		AccountName: account.Name,
		Public:      &Public{},
		account:     account,
		project:     p,
	}
	if len(p.Stages) == 0 {
		stage.Default = true
	}
	p.Stages = append(p.Stages, stage)
	return stage, nil
}

func (p *Project) RemoveStage(stageName string) {
	for idx, s := range p.Stages {
		if s.Name == stageName {
			p.Stages = append(p.Stages[:idx], p.Stages[idx+1:]...)
		}
	}
	p.setDefaultStage()
}

type EnvironmentConfig struct {
	Project ProjectEnvironmentConfig `yaml:"project"`
}

type ProjectEnvironmentConfig struct {
	Stages                []StageEnvironmentConfig `yaml:"stages"`
	FunctionConfiguration `yaml:",inline"`
}

type StageEnvironmentConfig struct {
	Name                  string                      `yaml:"name"`
	Functions             []FunctionEnvironmentConfig `yaml:"functions"`
	FunctionConfiguration `yaml:",inline"`
}

type FunctionEnvironmentConfig struct {
	Name                  string `yaml:"name"`
	FunctionConfiguration `yaml:",inline"`
}

const environmentConfigExample = `# Here you can define various configuration parameters
# for functions such as environment variables, memory size and timeout duration.
# These can be defined on a project, stage or function level. If the same parameter is
# defined on multiple levels the lowest level will take precedence.
# For example, uncommenting the config below will result in
# the function ping having the following configuration:
# memory_size: 512
# timeout: 60
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
#         env:
#           KEY3: function
`
