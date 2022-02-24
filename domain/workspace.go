package domain

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mantil-io/mantil/kit/token"
	"github.com/pkg/errors"
)

const (
	DefaultNodeName  = "dev"
	DefaultStageName = "dev"

	EnvWorkspace     = "MANTIL_WORKSPACE"
	EnvKey           = "MANTIL_KEY"
	EnvProjectName   = "MANTIL_PROJECT"
	EnvStageName     = "MANTIL_STAGE"
	EnvApiURL        = "MANTIL_API_URL"
	EnvSSMPathPrefix = "MANTIL_SSM_PATH_PREFIX"
	EnvKVTable       = "MANTIL_KV_TABLE"

	SSMPublicKey  = "public_key"
	SSMPrivateKey = "private_key"

	TagWorkspace   = EnvWorkspace
	TagKey         = EnvKey
	TagProjectName = EnvProjectName
	TagStageName   = EnvStageName

	// set to non empty to disable sending events from cli
	EnvNoEvents = "MANTIL_NO_EVENTS"
)

type Workspace struct {
	ID        string              `yaml:"id"`
	Version   string              `yaml:"version"`
	CreatedAt int64               `yaml:"created_at"`
	Projects  []*WorkspaceProject `yaml:"projects,omitempty"`
	Nodes     []*Node             `yaml:"nodes"`
}

type WorkspaceProject struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

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

	GitHubAuthEnabled bool   `yaml:"github_auth_enabled,omitempty"`
	JWT               string `yaml:"jwt,omitempty"`

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

func newWorkspace() *Workspace {
	return &Workspace{
		ID:        UID(),
		CreatedAt: time.Now().UnixMilli(),
	}
}

func (w *Workspace) RemoveNode(name string) {
	for idx, a := range w.Nodes {
		if a.Name == name {
			w.Nodes = append(w.Nodes[:idx], w.Nodes[idx+1:]...)
			return
		}
	}
}

func (w *Workspace) Node(name string) *Node {
	for _, a := range w.Nodes {
		if a.Name == name {
			return a
		}
	}
	return nil
}

func (w *Workspace) NewNode(name, awsAccountID, awsRegion, functionsBucket, functionsPath, version string, githubAuth bool) (*Node, error) {
	if w.nodeExists(name) {
		return nil, errors.WithStack(&NodeExistsError{name})
	}
	uid := uid4()
	bucket := fmt.Sprintf("mantil-%s", uid)
	a := &Node{
		Name:      name,
		ID:        uid,
		Version:   version,
		AccountID: awsAccountID,
		Region:    awsRegion,
		Bucket:    bucket,
		Functions: NodeFunctions{
			Bucket: functionsBucket,
			Path:   functionsPath,
		},
		workspace: w,
	}
	if githubAuth {
		a.GitHubAuthEnabled = true
	} else {
		publicKey, privateKey, err := token.KeyPair()
		if err != nil {
			return nil, errors.Wrap(err, "could not create public/private key pair")
		}
		a.Keys = NodeKeys{
			Public:  publicKey,
			Private: privateKey,
		}
	}
	w.Nodes = append(w.Nodes, a)
	return a, nil
}

func (w *Workspace) nodeExists(name string) bool {
	for _, n := range w.Nodes {
		if n.Name == name {
			return true
		}
	}
	return false
}

func (n *Node) ResourceTags() map[string]string {
	return map[string]string{
		TagWorkspace: n.workspace.ID,
		TagKey:       n.ID,
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
		EnvKVTable:       fmt.Sprintf("mantil-kv-%s", n.ID),
		EnvSSMPathPrefix: fmt.Sprintf("/mantil-node-%s", n.ID),
	}
}

func (w *Workspace) afterRestore() {
	for _, n := range w.Nodes {
		n.workspace = w
	}
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

// idea stolen from:  https://github.com/nats-io/nats-server/blob/fd9e9480dad9498ed8109e659fc8ed5c9b2a1b41/server/nkey.go#L41
// 4 byte uid hex encoded
func uid4() string {
	var rndData [4]byte
	data := rndData[:]
	_, _ = io.ReadFull(rand.Reader, data)
	encoded := make([]byte, hex.EncodedLen(len(data)))
	hex.Encode(encoded, data)
	return strings.ToLower(string(encoded[:]))
}

// 16 byte (guid) base64 encoded
func UID() string {
	buf := make([]byte, 22)
	uid := [16]byte(uuid.New())
	base64.RawURLEncoding.Encode(buf, uid[:])
	return string(buf)
}

func (w *Workspace) FindNode(name string) *Node {
	if name == "" && len(w.Nodes) == 1 {
		return w.Nodes[0]
	}
	return w.Node(name)
}

func (w *Workspace) NodeNames() []string {
	var names []string
	for _, n := range w.Nodes {
		names = append(names, n.Name)
	}
	return names
}

// Factory proper object model
func Factory(w *Workspace, p *Project, e *EnvironmentConfig) error {
	w.afterRestore()
	if p == nil {
		return nil
	}
	p.workspace = w
	p.environment = e

	for _, stage := range p.Stages {
		stage.project = p
		stage.node = w.Node(stage.NodeName)
		if stage.node == nil {
			return fmt.Errorf("can't find node %s for stage %s", stage.NodeName, stage.Name)
		}
		for _, f := range stage.Functions {
			f.stage = stage
		}
	}
	return nil
}

func (n *Node) AuthToken() (string, error) {
	if !n.GitHubAuthEnabled {
		claims := &AccessTokenClaims{
			Workspace: n.workspace.ID,
		}
		return token.JWT(n.Keys.Private, claims, 7*24*time.Hour)
	}
	t := n.JWT
	if t == "" {
		return "", &TokenExpiredError{}
	}
	exp, err := token.ExpiresAt(t)
	if err != nil {
		return "", err
	}
	if exp.Before(time.Now()) {
		return "", &TokenExpiredError{}
	}
	return t, nil
}

func (w *Workspace) AddProject(name, path string) {
	w.Projects = append(w.Projects, &WorkspaceProject{
		Name: name,
		Path: path,
	})
}

func (w *Workspace) RemoveProject(name string) {
	for idx, p := range w.Projects {
		if p.Name == name {
			w.Projects = append(w.Projects[:idx], w.Projects[idx+1:]...)
			return
		}
	}
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

func SSMParameterPath(key string) (string, error) {
	p, ok := os.LookupEnv(EnvSSMPathPrefix)
	if !ok {
		return "", &SSMPathNotFoundError{}
	}
	return fmt.Sprintf("%s/%s", p, key), nil
}
