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

	SSMPublicKey     = "public_key"
	SSMPrivateKey    = "private_key"
	SSMGithubUserKey = "github_user"

	NodeConfigKey = "config"

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
	NodeStore NodeStore           `yaml:"node_store,omitempty"`
}

type WorkspaceProject struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

func newWorkspace() *Workspace {
	return &Workspace{
		ID:        UID(),
		CreatedAt: time.Now().UnixMilli(),
	}
}

func (w *Workspace) AddNodeToken(token string) error {
	return w.NodeStore.UpsertNodeToken(token)
}

func (w *Workspace) AddNode(n *Node) {
	if w.nodeExists(n.Name) {
		return
	}
	w.Nodes = append(w.Nodes, n)
}

func (w *Workspace) RemoveNode(name string) {
	w.NodeStore.RemoveNode(name)
	for idx, a := range w.Nodes {
		if a.Name == name {
			w.Nodes = append(w.Nodes[:idx], w.Nodes[idx+1:]...)
			return
		}
	}
}

func (w *Workspace) Node(name string) *Node {
	n, _ := w.NodeStore.Node(name)
	if n != nil {
		return n
	}
	for _, a := range w.Nodes {
		if a.Name == name {
			return a
		}
	}
	return nil
}

func (w *Workspace) NewNode(name, awsAccountID, awsRegion, functionsBucket, functionsPath, version string, githubUser string) (*Node, error) {
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
	if githubUser != "" {
		a.GithubUser = githubUser
	} else {
		publicKey, privateKey, err := token.KeyPair()
		if err != nil {
			return nil, errors.Wrap(err, "could not create public/private key pair")
		}
		a.Keys = NodeKeys{
			Public:  publicKey,
			Private: privateKey,
		}
		w.Nodes = append(w.Nodes, a)
	}
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

func (w *Workspace) afterRestore() {
	for _, n := range w.Nodes {
		n.workspace = w
	}
	w.NodeStore.workspace = w
	w.NodeStore.afterRestore()
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
	n, _ := w.NodeStore.FindNode(name)
	if n != nil {
		return n
	}
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
	for _, n := range w.NodeStore.Nodes {
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

func SSMParameterPath(key string) (string, error) {
	p, ok := os.LookupEnv(EnvSSMPathPrefix)
	if !ok {
		return "", &SSMPathNotFoundError{}
	}
	return fmt.Sprintf("%s/%s", p, key), nil
}
