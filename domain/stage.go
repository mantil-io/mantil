package domain

import (
	"fmt"
)

const (
	EnvStageWsForwarder       = "MANTIL_STAGE_WS_FORWARDER"
	StateBucketPrefix         = "state"
	FunctionsBucketPrefix     = "functions"
	FunctionsBucketExpireDays = 7
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

func (s *Stage) FunctionsBucketPrefix() string {
	return fmt.Sprintf("%s/%s/%s", FunctionsBucketPrefix, s.project.Name, s.Name)
}

func (s *Stage) StateBucketPrefix() string {
	return fmt.Sprintf("%s/%s/%s", StateBucketPrefix, s.project.Name, s.Name)
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

func (s *Stage) applyConfiguration(ec *EnvironmentConfig) bool {
	if ec == nil {
		return false
	}
	sec := ec.Project.StageEnvConfig(s.Name)
	changed := false
	for _, f := range s.Functions {
		// ordered by priority from lowest to highest
		sources := []FunctionConfiguration{
			s.defaultFunctionConfiguration(),
			ec.Project.FunctionConfiguration,
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
	return map[string]string{
		EnvProjectName: s.project.Name,
		EnvStageName:   s.Name,
		EnvKey:         s.account.ResourceSuffix(),
		EnvSDKConfig:   s.sdkConfigEnv(),
	}
}

func (s *Stage) sdkConfigEnv() string {
	c := &SDKConfig{
		ResourceTags:    s.ResourceTags(),
		WsForwarderName: s.WsForwarderLambdaName(),
	}
	return c.Encode()
}

func (s *Stage) WsEnv() map[string]string {
	return map[string]string{
		EnvProjectName: s.project.Name,
		EnvStageName:   s.Name,
		EnvKey:         s.account.ResourceSuffix(),
	}
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

func (s *Stage) AddPublicSites(names []string) {
	for _, n := range names {
		s.Public.Sites = append(s.Public.Sites, &PublicSite{
			Name: n,
		})
	}
}

func (s *Stage) RemovePublicSites(names []string) {
	for _, n := range names {
		for idx, ps := range s.Public.Sites {
			if ps.Name == n {
				s.Public.Sites = append(s.Public.Sites[:idx], s.Public.Sites[idx+1:]...)
			}
		}
	}
}

func (s *Stage) PublicSites() []*PublicSite {
	if s.Public == nil {
		return nil
	}
	return s.Public.Sites
}

func (s *Stage) PublicSiteNames() []string {
	if s.Public == nil {
		return nil
	}
	var names []string
	for _, ps := range s.Public.Sites {
		names = append(names, ps.Name)
	}
	return names
}

func (s *Stage) WsForwarderLambdaName() string {
	return fmt.Sprintf("%s-%s-ws-forwarder-%s", s.project.Name, s.Name, s.account.ResourceSuffix())
}
