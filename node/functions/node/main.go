package main

import (
	"context"
	"log"

	"github.com/mantil-io/mantil.go"
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
	return n.store.StoreUser(req.Username)
}

type AddProjectRequest struct {
	Repo string `json:"repo"`
}

func (n *Node) AddProject(ctx context.Context, req *AddProjectRequest) error {
	return n.store.StoreProject(req.Repo)
}

func main() {
	var api = New()
	mantil.LambdaHandler(api)
}
