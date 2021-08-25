package deploy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil-backend/internal/deploy"
	"github.com/mantil-io/mantil-backend/internal/mantil"
)

type Deploy struct{}

type DeployRequest struct {
	ProjectName     string
	Token           string
	FunctionUpdates []mantil.FunctionUpdate
}
type DeployResponse struct {
	ApiURL string
}

func (h *Deploy) Init(ctx context.Context) {}

func (h *Deploy) Invoke(ctx context.Context, req *DeployRequest) (*DeployResponse, error) {
	return h.Deploy(ctx, req)
}

func (h *Deploy) Deploy(ctx context.Context, req *DeployRequest) (*DeployResponse, error) {
	if req.ProjectName == "" || req.Token == "" {
		return nil, fmt.Errorf("bad request")
	}
	p, err := mantil.LoadProject(req.ProjectName)
	if err != nil {
		return nil, err
	}
	if p.Token != req.Token {
		return nil, fmt.Errorf("access denied")
	}
	d, err := deploy.New(p, req.FunctionUpdates, "/tmp")
	if err != nil {
		return nil, err
	}
	err = d.Deploy()
	if err != nil {
		return nil, err
	}
	rsp := DeployResponse{
		ApiURL: p.ApiURL,
	}
	return &rsp, nil
}

func New() *Deploy {
	return &Deploy{}
}
