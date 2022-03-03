package main

import (
	"context"
	"log"

	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/api/node"
	"github.com/mantil-io/mantil/node/dto"
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

func (n *Node) AddUser(ctx context.Context, req *dto.AddUserRequest) error {
	ok, err := domain.IsOwner(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotAuthorized
	}
	return n.store.StoreUser(req.Username, req.Role)
}

func main() {
	var api = New()
	mantil.LambdaHandler(api)
}
