package deploy

import (
	"context"
	"fmt"
	"log"

	"github.com/atoz-technology/mantil-backend/internal/deploy"
	"github.com/atoz-technology/mantil-backend/internal/mantil"
	"github.com/atoz-technology/mantil-backend/internal/stream"
)

type Deploy struct{}

type DeployRequest struct {
	ProjectName     string
	Token           string
	FunctionUpdates []mantil.FunctionUpdate
}
type DeployResponse struct {
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
	err = stream.LambdaLogStream(ctx, d.Deploy)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rsp := DeployResponse{}
	return &rsp, nil
}

func New() *Deploy {
	return &Deploy{}
}
