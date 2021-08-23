package setup

import (
	"context"

	"github.com/mantil-io/mantil-backend/internal/setup"
)

type Setup struct{}

type SetupRequest struct {
	Destroy bool
}

type SetupResponse struct {
	APIGatewayURL string
}

func (f *Setup) Invoke(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
	return f.Setup(ctx, req)
}

func (f *Setup) Setup(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
	apiGatewayUrl, err := setup.Setup("/tmp", req.Destroy)
	if err != nil {
		return nil, err
	}
	return &SetupResponse{
		APIGatewayURL: apiGatewayUrl,
	}, nil
}

func New() *Setup {
	return &Setup{}
}
