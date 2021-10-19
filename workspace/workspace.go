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
	EnvKey = "MANTIL_KEY"
)

const (
	TagWorkspace = "MANTIL_WORKSPACE"
	TagKey       = "MANTIL_KEY"
)

var (
	ErrAccountExists     = fmt.Errorf("account already exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrProjectNotFound   = fmt.Errorf("no Mantil project found")
	ErrAccountNotFound   = fmt.Errorf("account not found")
	ErrStageExists       = fmt.Errorf("stage already exists")
)

type Workspace struct {
	Name     string     `yaml:"name"`
	Accounts []*Account `yaml:"accounts"`
}

type Account struct {
	Name      string           `yaml:"name"`
	ID        string           `yaml:"id"`
	UID       string           `yaml:"uid"`
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
	uid := uid()
	bucket := fmt.Sprintf("mantil-%s-%s", awsRegion, uid)
	a := &Account{
		Name:   name,
		ID:     awsAccountID,
		UID:    uid,
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

func (a *Account) ResourceTags() map[string]string {
	return map[string]string{
		TagWorkspace: a.workspace.Name,
		TagKey:       a.UID,
	}
}

// TODO: global out of domain model
func BackendLambdaFunctionName(functionName, suffix string) string {
	return fmt.Sprintf("mantil-%s-%s", functionName, suffix)
}

func (w *Workspace) Empty() bool {
	return len(w.Accounts) == 0
}

func (w *Workspace) afterRestore() {
	for _, a := range w.Accounts {
		a.workspace = w
	}
}

const (
	accountResourcePrefix = "mantil-setup"
)

func (a *Account) ResourceSuffix() string {
	return a.UID
}

func (a *Account) WorkspaceName() string {
	return a.workspace.Name
}

func (a *Account) SetupStackName() string {
	return a.SetupLambdaName()
}

func (a *Account) SetupLambdaName() string {
	return fmt.Sprintf("%s-%s", accountResourcePrefix, a.UID)
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
		return "workspace"
	}
	return strings.ToLower(u.Username)
}

func (w *Workspace) FindAccount(name string) *Account {
	if name == "" && len(w.Accounts) == 1 {
		return w.Accounts[0]
	}
	return w.Account(name)
}

func (w *Workspace) AccountNames() []string {
	var names []string
	for _, a := range w.Accounts {
		names = append(names, a.Name)
	}
	return names
}

// factory proper object model
func factory(w *Workspace, p *Project, e *EnvironmentConfig) error {
	w.afterRestore()
	if p == nil {
		return nil
	}
	p.workspace = w
	p.environment = e

	for _, stage := range p.Stages {
		stage.project = p
		stage.account = w.Account(stage.AccountName)
		if stage.account == nil {
			return fmt.Errorf("can't find account %s for stage %s", stage.AccountName, stage.Name)
		}
		for _, f := range stage.Functions {
			f.stage = stage
		}
	}
	return nil
}
