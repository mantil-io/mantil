package workspace

import (
	"fmt"
	"strings"
)

const (
	DefaultStageName = "dev"
	// env variable containing comma separated list of env variables
	// that should be added as tags to all resource created by stage lambda functions
	EnvMantilStageTags = "MANTIL_ENV_TAGS"
	TagStageName       = "MANTIL_STAGE"
)

var (
	// list of env variables for EnvMantilStageTags
	// variables specified here should be passed to all lambda functions
	MantilStageTags = []string{EnvProjectName, EnvStageName}
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
	return map[string]string{
		TagStageName: s.Name,
	}
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

func (s *Stage) LogGroupPrefix() string {
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

func (s *Stage) ApplyEnv() bool {
	ec := s.project.environment

	changed := false
	for _, f := range s.Functions {
		envChain := []map[string]string{
			s.defaultEnv(),
			ec.Project.Env,
		}
		for _, sc := range ec.Project.Stages {
			if sc.Name != s.Name {
				continue
			}
			envChain = append(envChain, sc.Env)
			for _, fc := range sc.Functions {
				if f.Name != fc.Name {
					continue
				}
				envChain = append(envChain, fc.Env)
			}
		}
		// reverse the chain to get correct priorities
		for i, j := 0, len(envChain)-1; i < j; i, j = i+1, j-1 {
			envChain[i], envChain[j] = envChain[j], envChain[i]
		}
		changed = f.mergeEnv(envChain...)
	}
	return changed
}

func (s *Stage) defaultEnv() map[string]string {
	env := map[string]string{
		EnvProjectName:     s.project.Name,
		EnvStageName:       s.Name,
		EnvWorkspaceKey:    s.Account().ResourceSuffix(),
		EnvMantilStageTags: strings.Join(MantilStageTags, ","),
	}
	return env
}

func (s *Stage) AddFunctions(names []string) *ErrReservedName {
	for _, name := range names {
		if !FunctionNameAvailable(name) {
			return &ErrReservedName{Name: name}
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
