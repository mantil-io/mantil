package initialize

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/internal/backend/api/initialize"
	"github.com/mantil-io/mantil/internal/mantil"
)

type Init struct{}

type InitRequest struct {
	ProjectName string
}

type InitResponse struct {
	Project *mantil.Project
}

func (f *Init) Invoke(ctx context.Context, req *InitRequest) (*InitResponse, error) {
	return f.Init(ctx, req)
}

func (f *Init) Init(ctx context.Context, req *InitRequest) (*InitResponse, error) {
	if !f.isRequestValid(req) {
		return nil, fmt.Errorf("bad request")
	}
	p, err := initialize.InitProject(req.ProjectName)
	if err != nil {
		return nil, err
	}
	return &InitResponse{
		Project: p,
	}, nil
}

func (f *Init) isRequestValid(req *InitRequest) bool {
	if req == nil {
		return false
	}
	return req.ProjectName != ""
}

func New() *Init {
	return &Init{}
}
