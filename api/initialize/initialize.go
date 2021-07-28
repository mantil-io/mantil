package initialize

import (
	"context"
	"log"

	"github.com/atoz-technology/mantil-backend/internal/initialize"
	"github.com/atoz-technology/mantil-backend/internal/stream"
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
	var token string
	err := stream.LambdaLogStream(ctx, func() error {
		var err error
		token, err = initialize.InitProject(req.ProjectName)
		return err
	})
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
