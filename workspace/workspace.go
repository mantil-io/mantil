package workspace

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io"
	"os/user"
	"strings"

	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/cli/log"
)

const (
	DefaultAccountName = "dev"
)

const (
	TagWorkspace = "MANTIL_WORKSPACE"
	TagKey       = "MANTIL_KEY"
)

const (
	EnvWorkspaceKey = "MANTIL_WORKSPACE_KEY"
)

var (
	ErrAccountExists     = fmt.Errorf("account already exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
)

type Workspace struct {
	Name     string     `yaml:"name"`
	UID      string     `yaml:"uid"`
	Accounts []*Account `yaml:"accounts"`
}

type Account struct {
	Name      string           `yaml:"name"`
	ID        string           `yaml:"id"`
	Region    string           `yaml:"region"`
	Bucket    string           `yaml:"bucket"`
	Keys      AccountKeys      `yaml:"keys"`
	Endpoints AccountEndpoints `yaml:"endpoints"`
	Functions AccountFunctions `yaml:"functions"`
	CliRole   string           `yaml:"cli_role"`
	workspace *Workspace
}

type AccountKeys struct {
	Public  string `yaml:"public"`
	Private string `yaml:"private"`
}

type AccountEndpoints struct {
	Rest string `yaml:"rest"`
	Ws   string `yaml:"ws"`
}

type AccountFunctions struct {
	Bucket string `yaml:"bucket"`
	Path   string `yaml:"key"`
}

func newWorkspace(name string) *Workspace {
	return &Workspace{
		Name: name,
		UID:  uid(),
	}
}

func (w *Workspace) RemoveAccount(name string) {
	for idx, a := range w.Accounts {
		if a.Name == name {
			w.Accounts = append(w.Accounts[:idx], w.Accounts[idx+1:]...)
			return
		}
	}
}

func (w *Workspace) Account(name string) *Account {
	for _, a := range w.Accounts {
		if a.Name == name {
			return a
		}
	}
	return nil
}

func (w *Workspace) NewAccount(name, awsAccountID, awsRegion, functionsBucket, functionsPath string) (*Account, error) {
	if w.accountExists(name) {
		return nil, ErrAccountExists
	}
	publicKey, privateKey, err := auth.CreateKeyPair()
	if err != nil {
		return nil, log.Wrap(err, "could not create public/private key pair")
	}
	bucket := fmt.Sprintf("mantil-%s-%s", awsRegion, w.ResourceSuffix())
	a := &Account{
		Name:   name,
		ID:     awsAccountID,
		Region: awsRegion,
		Bucket: bucket,
		Keys: AccountKeys{
			Public:  publicKey,
			Private: privateKey,
		},
		Functions: AccountFunctions{
			Bucket: functionsBucket,
			Path:   functionsPath,
		},
		workspace: w,
	}
	w.Accounts = append(w.Accounts, a)
	return a, nil
}

func (w *Workspace) accountExists(name string) bool {
	for _, a := range w.Accounts {
		if a.Name == name {
			return true
		}
	}
	return false
}

func (w *Workspace) ResourceName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, w.UID)
}

func (w *Workspace) ResourceTags() map[string]string {
	return map[string]string{
		TagWorkspace: w.Name,
		TagKey:       w.UID,
	}
}

func (w *Workspace) ResourceSuffix() string {
	return w.UID
}

func (w *Workspace) Empty() bool {
	return len(w.Accounts) == 0
}

func (w *Workspace) afterRestore() {
	for _, a := range w.Accounts {
		a.workspace = w
	}
	// when restored from previous version
	if w.UID == "" && len(w.Accounts) == 0 {
		w.UID = uid()
	}
}

func (a *Account) ResourceSuffix() string {
	return a.workspace.UID
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
	u, _ := user.Current()
	if u == nil {
		return ""
	}
	return strings.ToLower(u.Username)
}
