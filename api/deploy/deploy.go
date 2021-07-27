package deploy

import (
	"context"
	"fmt"

	"github.com/atoz-technology/mantil-backend/internal/deploy"
	"github.com/atoz-technology/mantil-backend/internal/mantil"
)

type Deploy struct{}

type DeployRequest struct {
	ProjectName     string
	Token           string
	FunctionUpdates []mantil.FunctionUpdate
}
type DeployResponse struct {
	Response string
}

func (h *Deploy) Init(ctx context.Context) {}

func (h *Deploy) Invoke(ctx context.Context, req *DeployRequest) (*DeployResponse, error) {
	return h.Deploy(ctx, req)
}

func (h *Deploy) Deploy(ctx context.Context, req *DeployRequest) (*DeployResponse, error) {
	if req.ProjectName == "" || req.Token == "" {
		return nil, fmt.Errorf("bad request")
	}
	projectBucket := mantil.ProjectBucket(req.ProjectName)
	p, err := mantil.LoadProject(projectBucket)
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
	if err := d.Deploy(); err != nil {
		return nil, err
	}
	rsp := DeployResponse{Response: "success"}
	return &rsp, nil
}

func New() *Deploy {
	return &Deploy{}
}
