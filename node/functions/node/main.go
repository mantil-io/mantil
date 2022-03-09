package main

import (
	"context"
	"log"
	"time"

	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/aws"
	"github.com/mantil-io/mantil/kit/token"
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
	ok, err := domain.IsAdmin(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotAuthorized
	}
	return n.store.StoreUser(req.Username, req.Role)
}

func (n *Node) AutomationJWT(ctx context.Context, req *dto.AutomationJWTRequest) (*dto.AutomationJWTResponse, error) {
	ok, err := domain.IsAdmin(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrNotAuthorized
	}
	node, err := n.store.FindConfig()
	if err != nil {
		return nil, err
	}
	token, err := n.generateAutomationJWT(req.Project, req.Stage, node)
	if err != nil {
		return nil, err
	}
	return &dto.AutomationJWTResponse{
		Token: token,
	}, nil
}

func (n *Node) generateAutomationJWT(project, stage string, node *domain.Node) (string, error) {
	privateKey, err := privateKey()
	if err != nil {
		return "", err
	}
	return token.JWT(privateKey, &domain.AccessTokenClaims{
		Role:    domain.Automation,
		Project: project,
		Stage:   stage,
		Node:    node,
	}, 365*24*time.Hour)
}

func privateKey() (string, error) {
	awsClient, err := aws.New()
	if err != nil {
		return "", err
	}
	path, err := domain.SSMParameterPath(domain.SSMPrivateKey)
	if err != nil {
		return "", err
	}
	return awsClient.GetSSMParameter(path)
}

func (n *Node) RemoveUser(ctx context.Context, req *dto.RemoveUserRequest) error {
	ok, err := domain.IsAdmin(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotAuthorized
	}
	return n.store.RemoveUser(req.Username)
}

func main() {
	var api = New()
	mantil.LambdaHandler(api)
}
