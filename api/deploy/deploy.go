package deploy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/internal/backend/api/deploy"
	"github.com/mantil-io/mantil/internal/backend/terraform"
	"github.com/mantil-io/mantil/internal/mantil"
)

type Deploy struct{}

type DeployRequest struct {
	ProjectName string
	Token       string
	Updates     []mantil.ProjectUpdate
}
type DeployResponse struct {
	Project *mantil.Project
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
	tf, err := terraform.New(req.ProjectName)
	if err != nil {
		return nil, err
	}
	defer tf.Cleanup()
	d, err := deploy.New(p, req.Updates, tf)
	if err != nil {
		return nil, err
	}
	err = d.Deploy()
	if err != nil {
		return nil, err
	}
	rsp := DeployResponse{
		Project: p,
	}
	return &rsp, nil
}

func New() *Deploy {
	return &Deploy{}
}
