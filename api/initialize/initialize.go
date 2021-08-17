package initialize

import (
	"context"
	"fmt"

	"github.com/atoz-technology/mantil-backend/internal/initialize"
)

type Init struct{}

type InitRequest struct {
	ProjectName string
}

type InitResponse struct {
	Token string
}

func (f *Init) Invoke(ctx context.Context, req *InitRequest) (*InitResponse, error) {
	return f.Init(ctx, req)
}

func (f *Init) Init(ctx context.Context, req *InitRequest) (*InitResponse, error) {
	if !f.isRequestValid(req) {
		return nil, fmt.Errorf("bad request")
	}
	token, err := initialize.InitProject(req.ProjectName)
	if err != nil {
		return nil, err
	}
	return &InitResponse{
		Token: token,
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
