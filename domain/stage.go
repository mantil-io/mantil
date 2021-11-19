package domain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
)

const (
	EnvMantilConfig           = "MANTIL_CONFIG"
	StateBucketPrefix         = "state"
	FunctionsBucketPrefix     = "functions"
	FunctionsBucketExpireDays = 7
)

type Stage struct {
	Name      string          `yaml:"name"`
	Default   bool            `yaml:"default,omitempty"`
	NodeName  string          `yaml:"node"`
	Endpoints *StageEndpoints `yaml:"endpoints,omitempty"`
	Functions []*Function     `yaml:"functions,omitempty"`
	Public    *Public         `yaml:"public,omitempty"`
	project   *Project
	node      *Node
}

type Public struct {
	Bucket string `yaml:"bucket"`
	Hash   string `yaml:"hash,omitempty"`
}

func (s *Stage) ResourceTags() map[string]string {
	// stage resource tags include tags from both node and project
	tags := s.node.ResourceTags()
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

func (s *Stage) Node() *Node {
	return s.node
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

func (s *Stage) ResourceNamingTemplate() string {
	prefix := fmt.Sprintf("%s-%s", s.project.Name, s.Name)
	suffix := s.node.UID
	return prefix + "-%s-" + suffix
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
		fc := f.FunctionConfiguration.merge(sources...)
		changed = changed || fc
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
		EnvKey:         s.node.ResourceSuffix(),
		EnvSDKConfig:   s.sdkConfigEnv(),
	}
}

func (s *Stage) sdkConfigEnv() string {
	c := &SDKConfig{
		ResourceTags:    s.ResourceTags(),
		WsForwarderName: s.WsForwarderLambdaName(),
		NamingTemplate:  s.ResourceNamingTemplate(),
	}
	return c.Encode()
}

func (s *Stage) WsEnv() map[string]string {
	return map[string]string{
		EnvMantilConfig: s.WsConfig().Encode(),
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

func (s *Stage) WsForwarderLambdaName() string {
	return fmt.Sprintf("%s-%s-ws-forwarder-%s", s.project.Name, s.Name, s.node.ResourceSuffix())
}

func (s *Stage) RestEndpoint() string {
	if s.Endpoints == nil {
		return ""
	}
	return s.Endpoints.Rest
}

func (s *Stage) WsEndpoint() string {
	if s.Endpoints == nil {
		return ""
	}
	return fmt.Sprintf("%s/$default", s.Endpoints.Ws)
}

type WsConfig struct {
	ApiToFn map[string]string `json:"apiToFn"`
}

func (c WsConfig) Encode() string {
	buf, _ := json.Marshal(c)
	return base64.StdEncoding.EncodeToString(buf)
}

func (s *Stage) WsConfig() WsConfig {
	apiToFn := map[string]string{}
	for _, f := range s.Functions {
		apiToFn[f.Name] = f.LambdaName()
	}
	return WsConfig{
		ApiToFn: apiToFn,
	}
}

func (s *Stage) PublicEnv() ([]byte, error) {
	data := struct {
		RestEndpoint string
		WsEndpoint   string
	}{
		RestEndpoint: s.RestEndpoint(),
		WsEndpoint:   s.WsEndpoint(),
	}
	tpl, err := template.New("").Parse(publicEnvTemplate)
	if err != nil {
		return nil, err
	}
	out := bytes.NewBuffer(nil)
	if err := tpl.Execute(out, data); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

const publicEnvTemplate = `var mantilEnv = {
	endpoints: {
		rest: '{{.RestEndpoint}}',
		ws: '{{.WsEndpoint}}',
	},
};
`

const PublicEnvKey = "mantil_env.js"

func (s *Stage) HasPublic() bool {
	return s.Public != nil
}
