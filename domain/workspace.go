package domain

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/mantil-io/mantil/kit/token"
	"github.com/pkg/errors"
)

const (
	DefaultNodeName  = "dev"
	DefaultStageName = "dev"

	EnvWorkspace   = "MANTIL_WORKSPACE"
	EnvKey         = "MANTIL_KEY"
	EnvProjectName = "MANTIL_PROJECT"
	EnvStageName   = "MANTIL_STAGE"
	EnvApiURL      = "MANTIL_API_URL"

	TagWorkspace   = EnvWorkspace
	TagKey         = EnvKey
	TagProjectName = EnvProjectName
	TagStageName   = EnvStageName

	configDir             = "config"
	configName            = "state.yml"
	environmentConfigName = "environment.yml"
)

type Workspace struct {
	Name  string  `yaml:"name"`
	Nodes []*Node `yaml:"nodes"`
}

type Node struct {
	Name      string        `yaml:"name"`
	ID        string        `yaml:"id"`
	UID       string        `yaml:"uid"`
	Region    string        `yaml:"region"`
	Bucket    string        `yaml:"bucket"`
	Keys      NodeKeys      `yaml:"keys"`
	Endpoints NodeEndpoints `yaml:"endpoints"`
	Functions NodeFunctions `yaml:"functions"`
	CliRole   string        `yaml:"cli_role"`
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

func newWorkspace(name string) *Workspace {
	return &Workspace{
		Name: name,
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

func (w *Workspace) NewNode(name, awsAccountID, awsRegion, functionsBucket, functionsPath string) (*Node, error) {
	if w.nodeExists(name) {
		return nil, errors.WithStack(&NodeExistsError{name})
	}
	publicKey, privateKey, err := token.KeyPair()
	if err != nil {
		return nil, errors.Wrap(err, "could not create public/private key pair")
	}
	uid := uid()
	bucket := fmt.Sprintf("mantil-%s-%s", awsRegion, uid)
	a := &Node{
		Name:   name,
		ID:     awsAccountID,
		UID:    uid,
		Region: awsRegion,
		Bucket: bucket,
		Keys: NodeKeys{
			Public:  publicKey,
			Private: privateKey,
		},
		Functions: NodeFunctions{
			Bucket: functionsBucket,
			Path:   functionsPath,
		},
		workspace: w,
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
		TagWorkspace: n.workspace.Name,
		TagKey:       n.UID,
	}
}

func (n *Node) AuthEnv() map[string]string {
	return map[string]string{
		EnvPublicKey: n.Keys.Public,
	}
}

func (w *Workspace) Empty() bool {
	return len(w.Nodes) == 0
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
	return n.UID
}

func (n *Node) WorkspaceName() string {
	return n.workspace.Name
}

func (n *Node) SetupStackName() string {
	return n.SetupLambdaName()
}

func (n *Node) SetupLambdaName() string {
	return fmt.Sprintf("%s-%s", nodeResourcePrefix, n.UID)
}

// idea stolen from:  https://github.com/nats-io/nats-server/blob/fd9e9480dad9498ed8109e659fc8ed5c9b2a1b41/server/nkey.go#L41
func uid() string {
	var rndData [4]byte
	data := rndData[:]
	_, _ = io.ReadFull(rand.Reader, data)
	var encoded [7]byte
	base32.StdEncoding.WithPadding(base32.NoPadding).Encode(encoded[:], data)
	return strings.ToLower(string(encoded[:]))
}

func defaultWorkspaceName() string {
	dflt := "workspace"

	u, _ := user.Current()
	if u == nil {
		return dflt
	}
	username := u.Username
	if strings.Contains(username, `\`) {
		parts := strings.Split(username, `\`)
		username = parts[len(parts)-1]
	}
	if username == "" {
		return dflt
	}
	username = strings.ToLower(username)

	// Make a Regex to say we only want letters and numbers
	reg := regexp.MustCompile("[^a-z0-9]+")
	username = reg.ReplaceAllString(username, "")

	if username == "" {
		return dflt
	}
	return username
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
	claims := &AccessTokenClaims{
		Workspace: n.WorkspaceName(),
	}
	return token.JWT(n.Keys.Private, claims, 7*24*time.Hour)
}
