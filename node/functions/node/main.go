package main

import (
	"context"
	"log"

	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/api/node"
)

type Node struct {
	store *node.Store
}

func New() *Node {
	s, err := node.NewStore()
	if err != nil {
		log.Fatal(err)
	}
	return &Node{
		store: s,
	}
}

type AddUserRequest struct {
	Username string `json:"username"`
}

func (n *Node) AddUser(ctx context.Context, req *AddUserRequest) error {
	ok, err := domain.IsOwner(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotAuthorized
	}
	return n.store.StoreUser(req.Username)
}

type AddProjectRequest struct {
	Name string `json:"name"`
	Repo string `json:"repo"`
}

func (n *Node) AddProject(ctx context.Context, req *AddProjectRequest) error {
	ok, err := domain.IsOwner(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotAuthorized
	}
	return n.store.StoreProject(req.Name, req.Repo)
}

func main() {
	var api = New()
	mantil.LambdaHandler(api)
}
