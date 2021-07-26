package deploy

import (
	"context"

	"github.com/atoz-technology/mantil-backend/internal/deploy"
	"github.com/atoz-technology/mantil-backend/internal/mantil"
)

type Deploy struct{}

type DeployRequest struct {
	ProjectBucket   string //TODO use mantil token here
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
	p, err := mantil.LoadProject(req.ProjectBucket)
	if err != nil {
		return nil, err
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
