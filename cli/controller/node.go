package controller

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/dto"
)

const (
	UserAddHTTPMethod    = "node/addUser"
	UserRemoveHTTPMethod = "node/removeUser"
	LoginHTTPMethod      = "auth/login"
)

type NodeUserAddArgs struct {
	Node       string
	GithubUser string
	Role       string
}

func NodeUserAdd(a NodeUserAddArgs) error {
	n, err := findNode(a.Node)
	if err != nil {
		return err
	}
	i, err := nodeInvoker(n)
	if err != nil {
		return err
	}
	r, err := resolveRole(a.Role)
	if err != nil {
		return err
	}
	if err := i.Do(UserAddHTTPMethod, &dto.AddUserRequest{
		Username: a.GithubUser,
		Role:     r,
	}, nil); err != nil {
		return err
	}
	ui.Info("Successfully added user %s. They can now login using the command `mantil node login %s`", a.GithubUser, n.Endpoints.Rest)
	return nil
}

type NodeUserRemoveArgs struct {
	Node     string
	Username string
}

func NodeUserRemove(a NodeUserRemoveArgs) error {
	n, err := findNode(a.Node)
	if err != nil {
		return err
	}
	i, err := nodeInvoker(n)
	if err != nil {
		return err
	}
	return i.Do(UserRemoveHTTPMethod, &dto.RemoveUserRequest{
		Username: a.Username,
	}, nil)
}

func findNode(name string) (*domain.Node, error) {
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	if err != nil {
		return nil, err
	}
	w := fs.Workspace()
	if len(w.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes avaiable")
	}
	n := w.FindNode(name)
	if n == nil {
		return nil, fmt.Errorf("node not found")
	}
	return n, nil
}

func resolveRole(r string) (domain.Role, error) {
	switch r {
	case "admin":
		return domain.Admin, nil
	case "user":
		return domain.User, nil
	default:
		return -1, fmt.Errorf("unknown role")
	}
}

type NodeLoginArgs struct {
	NodeURL string
}

func NodeLogin(a NodeLoginArgs) error {
	t, err := githubAuth(a.NodeURL)
	if err != nil {
		return err
	}
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	if err != nil {
		return err
	}
	w := fs.Workspace()
	w.AddNodeToken(t)
	return fs.Store()
}

type NodeLogoutArgs struct {
	NodeName string
}

func NodeLogout(a NodeLogoutArgs) error {
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	if err != nil {
		return err
	}
	w := fs.Workspace()
	w.RemoveNode(a.NodeName)
	return fs.Store()
}
