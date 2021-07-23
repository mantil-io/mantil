package initialize

import (
	"context"
	"log"

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
	if req == nil {
		return nil, nil
	}
	token, err := initialize.InitProject(req.ProjectName)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	return &InitResponse{
		Token: token,
	}, nil
}

func New() *Init {
	return &Init{}
}
