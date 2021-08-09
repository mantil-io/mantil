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
	APIGatewayURL string
}

func (f *Bootstrap) Invoke(ctx context.Context, req *BootstrapRequest) (*BootstrapResponse, error) {
	return f.Bootstrap(ctx, req)
}

func (f *Bootstrap) Bootstrap(ctx context.Context, req *BootstrapRequest) (*BootstrapResponse, error) {
	var apiGatewayUrl string
	if err := stream.LambdaLogStream(ctx, func() error {
		url, err := bootstrap.Bootstrap("/tmp", req.Destroy)
		if err != nil {
			return err
		}
		apiGatewayUrl = url
		return nil
	}); err != nil {
		log.Println(err)
		return nil, err
	}
	return &BootstrapResponse{
		APIGatewayURL: apiGatewayUrl,
	}, nil
}

func New() *Bootstrap {
	return &Bootstrap{}
}
