package workspace

import (
	"fmt"
)

type Stage struct {
	Name        string          `yaml:"name"`
	Default     bool            `yaml:"default,omitempty"`
	AccountName string          `yaml:"account"`
	Endpoints   *StageEndpoints `yaml:"endpoints,omitempty"`
	Functions   []*Function     `yaml:"functions,omitempty"`
	Public      *Public         `yaml:"public,omitempty"`
	project     *Project
	account     *Account
}

type Public struct {
	Bucket string        `yaml:"bucket"`
	Sites  []*PublicSite `yaml:"sites"`
}

type PublicSite struct {
	Name string `yaml:"name"`
	Hash string `yaml:"hash"`
}

func (s *Stage) ResourceTags() map[string]string {
	// stage resource tags include tags from both account and project
	tags := s.account.ResourceTags()
	for k, v := range s.project.resourceTags() {
		tags[k] = v
	}
	tags[TagStageName] = s.Name
	return tags
}

type StageEndpoints struct {
	Rest string `yaml:"rest"`
	Ws   string `yaml:"ws"`
}

func (s *Stage) Account() *Account {
	return s.account
}

func (s *Stage) Project() *Project {
	return s.project
}

func (s *Stage) BucketPrefix() string {
	return fmt.Sprintf("stages/%s/%s", s.project.Name, s.Name)
}

func (s *Stage) LogGroupsPrefix() string {
	return fmt.Sprintf("%s-%s", s.project.Name, s.Name)
}

func (s *Stage) SetPublicBucket(bucket string) {
	if s.Public == nil {
		s.Public = &Public{}
	}
	s.Public.Bucket = bucket
}

func (s *Stage) SetEndpoints(rest, ws string) {
	s.Endpoints = &StageEndpoints{
		Rest: rest,
		Ws:   ws,
	}
}

func (s *Stage) ApplyConfiguration() bool {

	changed := false
	pe := s.project.environment
	sec := pe.Project.StageEnvConfig(s.Name)
	for _, f := range s.Functions {
		// ordered by priority from lowest to highest
		sources := []FunctionConfiguration{
			s.defaultFunctionConfiguration(),
			pe.Project.FunctionConfiguration,
			sec.FunctionConfiguration,
			sec.FunctionEnvConfig(f.Name).FunctionConfiguration,
		}
		changed = f.FunctionConfiguration.merge(sources...)
	}
	return changed
}

func (s *Stage) defaultFunctionConfiguration() FunctionConfiguration {
	return FunctionConfiguration{
		MemorySize: 128,
		Timeout:    60 * 15,
		Env:        s.defaultEnv(),
	}
}

func (s *Stage) defaultEnv() map[string]string {
	// default env includes resources tags as a way to communicate
	// to functions which tags need to be added to dynamically created resources
	return s.ResourceTags()
}

func (s *Stage) AddFunctions(names []string) error {
	for _, name := range names {
		if !FunctionNameAvailable(name) {
			return &ErrReservedName{Name: name}
		}
		if err := ValidateName(name); err != nil {
			return err
		}
		s.addFunction(name)
	}
	return nil
}

func (s *Stage) addFunction(name string) {
	f := &Function{
		Name:  name,
		stage: s,
	}
	f.addDefaults()
	s.Functions = append(s.Functions, f)
}

func (s *Stage) RemoveFunctions(removed []string) {
	for _, r := range removed {
		for idx, sf := range s.Functions {
			if sf.Name == r {
				s.Functions = append(s.Functions[:idx], s.Functions[idx+1:]...)
				break
			}
		}
	}
}

func (s *Stage) FunctionNames() []string {
	var names []string
	for _, f := range s.Functions {
		names = append(names, f.Name)
	}
	return names
}

func (s *Stage) FindFunction(name string) *Function {
	for _, f := range s.Functions {
		if f.Name == name {
			return f
		}
	}
	return nil
}
