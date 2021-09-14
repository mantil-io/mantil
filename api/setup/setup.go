package setup

import (
	"context"

	"github.com/mantil-io/mantil/internal/backend/api/setup"
	"github.com/mantil-io/mantil/internal/backend/terraform"
)

type Setup struct{}

type SetupRequest struct {
	Destroy bool
}

type SetupResponse struct {
	APIGatewayRestURL string
	APIGatewayWsURL   string
}

func (f *Setup) Invoke(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
	return f.Setup(ctx, req)
}

func (f *Setup) Setup(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
	tf, err := terraform.New("mantil-setup")
	if err != nil {
		return nil, err
	}
	defer tf.Cleanup()
	out, err := setup.Setup(tf, req.Destroy)
	if err != nil {
		return nil, err
	}
	return &SetupResponse{
		APIGatewayRestURL: out.RestURL,
		APIGatewayWsURL:   out.WsURL,
	}, nil
}

func New() *Setup {
	return &Setup{}
}
