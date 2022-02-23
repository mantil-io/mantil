package domain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"strings"
	"time"

	"github.com/mantil-io/mantil/kit/token"
)

const (
	EnvMantilConfig = "MANTIL_CONFIG"
)

type Stage struct {
	Name           string          `yaml:"name"`
	Default        bool            `yaml:"default,omitempty"`
	NodeName       string          `yaml:"node"`
	Keys           StageKeys       `yaml:"keys"`
	Endpoints      *StageEndpoints `yaml:"endpoints,omitempty"`
	LastDeployment *LastDeployment `yaml:"last_deployment,omitempty"`
	Functions      []*Function     `yaml:"functions,omitempty"`
	Public         *Public         `yaml:"public,omitempty"`
	CustomDomain   CustomDomain    `yaml:"custom_domain,omitempty"`
	project        *Project
	node           *Node
}

type StageKeys struct {
	Public  string `yaml:"public"`
	Private string `yaml:"private"`
}

type Public struct {
	Bucket string `yaml:"bucket"`
	Hash   string `yaml:"hash,omitempty"`
}

type LastDeployment struct {
	Version   string `yaml:"version"`
	Timestamp int64  `yaml:"timestamp"`
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
	return fmt.Sprintf("%s/%s", s.project.FunctionsBucketPrefix(), s.Name)
}

func (s *Stage) StateBucketPrefix() string {
	return fmt.Sprintf("%s/%s", s.project.StateBucketPrefix(), s.Name)
}

func (s *Stage) BucketPrefixes() []string {
	return []string{s.FunctionsBucketPrefix(), s.StateBucketPrefix()}
}

func (s *Stage) LogGroupsPrefix() string {
	return fmt.Sprintf("%s-%s", s.project.Name, s.Name)
}

func (s *Stage) ResourceNamingTemplate() string {
	prefix := fmt.Sprintf("%s-%s", s.project.Name, s.Name)
	suffix := s.node.ID
	return prefix + "-%s-" + suffix
}

func (s *Stage) AuthToken() (string, error) {
	claims := &AccessTokenClaims{
		Workspace: s.node.workspace.ID,
		Project:   s.project.Name,
		Stage:     s.Name,
	}
	return token.JWT(s.Keys.Private, claims, 7*24*time.Hour)
}

func (s *Stage) AuthEnv() map[string]string {
	return map[string]string{
		EnvPublicKey: s.Keys.Public,
	}
}

func (s Stage) mantilResourceNamingTemplate() string {
	return fmt.Sprintf("mantil-%s-%s", s.project.Name, s.Name) +
		"-%s-" +
		s.node.ResourceSuffix()
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

func (s *Stage) SetLastDeployment() {
	s.LastDeployment = &LastDeployment{
		Version:   s.node.Version,
		Timestamp: time.Now().UnixMilli(),
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
	if !reflect.DeepEqual(s.CustomDomain, sec.CustomDomain) {
		s.CustomDomain = sec.CustomDomain
		s.CustomDomain.setDefaults()
		changed = true
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
	return fmt.Sprintf(s.mantilResourceNamingTemplate(), "ws-forwarder")
}

func (s *Stage) RestEndpoint() string {
	if s.CustomDomain.DomainName != "" {
		d := s.CustomDomain.DomainName
		if s.CustomDomain.HttpSubdomain != "" {
			d = fmt.Sprintf("%s.%s", s.CustomDomain.HttpSubdomain, d)
		}
		return fmt.Sprintf("https://%s", d)
	}
	if s.Endpoints != nil {
		return s.Endpoints.Rest
	}
	return ""
}

func (s *Stage) WsEndpoint() string {
	if s.CustomDomain.DomainName != "" {
		return fmt.Sprintf("wss://%s.%s", s.CustomDomain.WsSubdomain, s.CustomDomain.DomainName)
	}
	if s.Endpoints != nil {
		return fmt.Sprintf("%s/$default", s.Endpoints.Ws)
	}
	return ""
}

func (s *Stage) PublicBucketName() string {
	return fmt.Sprintf("%s-%s-public-%s", s.Project().Name, s.Name, s.Node().ResourceSuffix())
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

func (s *Stage) AsCliStage() *CliStage {
	if s == nil {
		return nil
	}
	return &CliStage{
		Name:      s.Name,
		Node:      s.NodeName,
		Functions: len(s.Functions),
	}
}

type AwsResource struct {
	Name    string
	AWSName string
	Type    string
}

func (r AwsResource) LogGroup() string {
	switch r.Type {
	case AwsResourceLambda:
		return fmt.Sprintf("/aws/lambda/%s", r.AWSName)
	case AwsResourceAPIGateway:
		// uf: insert access-logs string before key
		a := strings.Split(r.AWSName, "-")
		name := strings.Join(append(a[:len(a)-1], "access-logs", a[len(a)-1]), "-")
		return fmt.Sprintf("/aws/vendedlogs/%s", name)
	default:
		return ""
	}
}

const (
	AwsResourceLambda     = "Lambda Function"
	AwsResourceAPIGateway = "API Gateway"
	AwsResourceS3Bucket   = "S3 Bucket"
	AwsResourceDynamoDB   = "DynamoDB Table"
	AwsResourceStack      = "CloudFormation Stack"
)

// Resources list of resources created for the stage
func (s *Stage) Resources() []AwsResource {
	var ar []AwsResource
	for _, f := range s.Functions {
		ar = append(ar, AwsResource{f.Name, f.LambdaName(), AwsResourceLambda})
	}
	ar = append(ar, AwsResource{"ws-forwarder", s.mantilResourceName("ws-forwarder"), AwsResourceLambda})
	ar = append(ar, AwsResource{"ws-handler", s.mantilResourceName("ws-handler"), AwsResourceLambda})
	ar = append(ar, AwsResource{"ws-connections", s.mantilResourceName("ws-connections"), AwsResourceDynamoDB})
	ar = append(ar, AwsResource{"kv", s.mantilResourceName("kv"), AwsResourceDynamoDB})

	ar = append(ar, AwsResource{"http", s.resourceName("http"), AwsResourceAPIGateway})
	ar = append(ar, AwsResource{"ws", s.resourceName("ws"), AwsResourceAPIGateway})

	if s.Public != nil {
		ar = append(ar, AwsResource{"", s.Public.Bucket, AwsResourceS3Bucket})
	}

	return ar
}

func (s *Stage) resourceName(name string) string {
	return fmt.Sprintf(s.ResourceNamingTemplate(), name)
}

func (s *Stage) mantilResourceName(name string) string {
	return fmt.Sprintf(s.mantilResourceNamingTemplate(), name)
}
