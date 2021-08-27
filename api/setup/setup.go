package setup

import (
	"context"

	"github.com/mantil-io/mantil-backend/internal/setup"
	"github.com/mantil-io/mantil-backend/internal/terraform"
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
	tf, err := terraform.New("mantil-setup")
	if err != nil {
		return nil, err
	}
	apiGatewayUrl, err := setup.Setup(tf, req.Destroy)
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
