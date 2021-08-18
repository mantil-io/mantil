package setup

import (
	"context"
	"log"

	"github.com/atoz-technology/mantil-backend/internal/setup"
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
		log.Println(err)
		return nil, err
	}
	return &SetupResponse{
		APIGatewayURL: apiGatewayUrl,
	}, nil
}

func New() *Setup {
	return &Setup{}
}
