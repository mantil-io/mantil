package main

import (
	"context"
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
	return &Node{}
}

func (n *Node) AddUser(ctx context.Context, req *dto.AddUserRequest) error {
	ok, err := domain.IsOwner(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrNotAuthorized
	}
	s, err := node.NewStore()
	if err != nil {
		return err
	}
	return s.StoreUser(req.Username, req.Role)
}

func (n *Node) AutomationJWT(ctx context.Context, req *dto.AutomationJWTRequest) (*dto.AutomationJWTResponse, error) {
	ok, err := domain.IsOwner(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, domain.ErrNotAuthorized
	}
	token, err := n.generateAutomationJWT(req.Project, req.Stage)
	if err != nil {
		return nil, err
	}
	return &dto.AutomationJWTResponse{
		Token: token,
	}, nil
}

func (n *Node) generateAutomationJWT(project, stage string) (string, error) {
	privateKey, err := privateKey()
	if err != nil {
		return "", err
	}
	return token.JWT(privateKey, &domain.AccessTokenClaims{
		Role:    domain.Automation,
		Project: project,
		Stage:   stage,
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

func main() {
	var api = New()
	mantil.LambdaHandler(api)
}
