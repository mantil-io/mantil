package bootstrap

import (
	"context"
	"log"

	"github.com/atoz-technology/mantil-backend/internal/bootstrap"
	"github.com/atoz-technology/mantil-backend/internal/stream"
)

type Bootstrap struct{}

type BootstrapRequest struct {
	Destroy bool
}

type BootstrapResponse struct {
}

func (f *Bootstrap) Invoke(ctx context.Context, req *BootstrapRequest) (*BootstrapResponse, error) {
	return f.Bootstrap(ctx, req)
}

func (f *Bootstrap) Bootstrap(ctx context.Context, req *BootstrapRequest) (*BootstrapResponse, error) {
	if err := stream.LambdaLogStream(ctx, func() error {
		return bootstrap.Bootstrap("/tmp", req.Destroy)
	}); err != nil {
		log.Println(err)
		return nil, err
	}
	return &BootstrapResponse{}, nil
}

func New() *Bootstrap {
	return &Bootstrap{}
}
