package domain

import (
	"fmt"
	"time"

	"github.com/mantil-io/mantil/kit/token"
)

type Node struct {
	Name    string `yaml:"name,omitempty"`
	ID      string `yaml:"id"`
	Version string `yaml:"version"`
	// AWS related attributes
	AccountID string `yaml:"accountID"` // AWS account id
	Region    string `yaml:"region"`    // AWS region
	Bucket    string `yaml:"bucket"`    // bucket name created on AWS
	CliRole   string `yaml:"cli_role"`  // role name for security node lambda function

	Keys      NodeKeys      `yaml:"keys,omitempty"`
	Endpoints NodeEndpoints `yaml:"endpoints"`
	Functions NodeFunctions `yaml:"functions"`
	Stages    []*NodeStage  `yaml:"stages,omitempty"`

	GithubUser string `yaml:"github_user,omitempty"`

	workspace *Workspace
}

type NodeKeys struct {
	Public  string `yaml:"public"`
	Private string `yaml:"private"`
}

type NodeEndpoints struct {
	Rest string `yaml:"rest"`
}

type NodeFunctions struct {
	Bucket string `yaml:"bucket"`
	Path   string `yaml:"key"`
}

type NodeStage struct {
	Name        string `yaml:"name"`
	ProjectName string `yaml:"project_name"`
}

func (n *Node) ResourceTags() map[string]string {
	return map[string]string{
		TagKey: n.ID,
	}
}

func (n *Node) UpgradeVersion(version, functionsBbucket, functionsPath string) {
	n.Version = version
	n.Functions.Bucket = functionsBbucket
	n.Functions.Path = functionsPath
}

func (n *Node) AuthEnv() map[string]string {
	return map[string]string{
		EnvPublicKey:     n.Keys.Public,
		EnvKVTable:       n.KVTableName(),
		EnvSSMPathPrefix: fmt.Sprintf("/mantil-node-%s", n.ID),
	}
}

func (n *Node) SetupEnv() map[string]string {
	return map[string]string{
		EnvKVTable: n.KVTableName(),
	}
}

func (n *Node) KVTableName() string {
	return fmt.Sprintf("mantil-kv-%s", n.ID)
}

const (
	nodeResourcePrefix = "mantil-setup"
)

func (n *Node) ResourceSuffix() string {
	return n.ID
}

func (n *Node) ResourceNamingTemplate() string {
	return "mantil-%s-" + n.ID
}

func (n *Node) SetupStackName() string {
	return n.SetupLambdaName()
}

func (n *Node) SetupLambdaName() string {
	return fmt.Sprintf("%s-%s", nodeResourcePrefix, n.ID)
}

func (n *Node) AuthToken() (string, error) {
	if !n.GithubAuthEnabled() {
		claims := &AccessTokenClaims{
			Role:      Admin,
			Workspace: n.workspace.ID,
		}
		return token.JWT(n.Keys.Private, claims, 7*24*time.Hour)
	}
	t := n.workspace.NodeStore.Token(n.Name)
	if t == "" {
		return "", &TokenExpiredError{}
	}
	exp, err := token.ExpiresIn(t)
	if err != nil {
		return "", err
	}
	if exp < 0 {
		return "", &TokenExpiredError{}
	}
	return t, nil
}

func (n *Node) UpdateToken(token string) {
	n.workspace.NodeStore.UpsertNodeToken(token)
}

func (n *Node) AddStage(name, projectName, path string) {
	n.Stages = append(n.Stages, &NodeStage{
		Name:        name,
		ProjectName: projectName,
	})
}

func (n *Node) RemoveStage(name string) {
	for idx, s := range n.Stages {
		if s.Name == name {
			n.Stages = append(n.Stages[:idx], n.Stages[idx+1:]...)
			return
		}
	}
}

func (n *Node) resourceName(name string) string {
	return fmt.Sprintf("mantil-%s-%s", name, n.ID)
}

func (n *Node) Resources() []AwsResource {
	var ar []AwsResource
	for _, name := range []string{"setup", "authorizer", "deploy", "destroy", "security"} {
		ar = append(ar, AwsResource{name, n.resourceName(name), AwsResourceLambda})
	}
	ar = append(ar, AwsResource{"setup", n.SetupStackName(), AwsResourceStack})
	ar = append(ar, AwsResource{"http", n.resourceName("http"), AwsResourceAPIGateway})
	ar = append(ar, AwsResource{"", n.Bucket, AwsResourceS3Bucket})

	return ar
}

func (n *Node) GithubAuthEnabled() bool {
	return n.GithubUser != ""
}

type NodeStore struct {
	Nodes     []*NodeStoreEntry `yaml:"nodes"`
	workspace *Workspace
}

type NodeStoreEntry struct {
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
	store *NodeStore
}

func (s *NodeStore) afterRestore() {
	for _, n := range s.Nodes {
		n.store = s
	}
}

func (s *NodeStore) Node(name string) (*Node, error) {
	var ne *NodeStoreEntry
	for _, n := range s.Nodes {
		if n.Name == name {
			ne = n
			break
		}
	}
	if ne == nil {
		return nil, &NodeNotFoundError{Name: name}
	}
	n, err := nodeFromToken(ne.Token)
	if err != nil {
		return nil, err
	}
	n.workspace = ne.store.workspace
	return n, nil
}

func (s *NodeStore) FindNode(name string) (*Node, error) {
	if name == "" && len(s.Nodes) == 1 {
		return s.Node(s.Nodes[0].Name)
	}
	return s.Node(name)
}

func (s *NodeStore) UpsertNodeToken(token string) error {
	n, err := nodeFromToken(token)
	if err != nil {
		return err
	}
	e := &NodeStoreEntry{
		Name:  n.Name,
		Token: token,
		store: s,
	}
	for idx, no := range s.Nodes {
		if no.Name == n.Name {
			s.Nodes[idx] = e
			return nil
		}
	}
	s.Nodes = append(s.Nodes, e)
	return nil
}

func nodeFromToken(t string) (*Node, error) {
	var claims AccessTokenClaims
	if err := token.Decode(t, &claims); err != nil {
		return nil, err
	}
	return claims.Node, nil
}

func (s *NodeStore) RemoveNode(name string) {
	for idx, n := range s.Nodes {
		if n.Name == name {
			s.Nodes = append(s.Nodes[:idx], s.Nodes[idx+1:]...)
			return
		}
	}
}

func (s *NodeStore) Token(nodeName string) string {
	for _, n := range s.Nodes {
		if n.Name == nodeName {
			return n.Token
		}
	}
	return ""
}
