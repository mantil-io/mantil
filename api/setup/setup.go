package setup

import (
	"context"
	"github.com/mantil-io/mantil/backend/api/setup"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/terraform"
)

type Setup struct{}

type SetupRequest struct {
	Version         string
	FunctionsBucket string
	FunctionsPath   string
	PublicKey       string
	Destroy         bool
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
	rc := &config.RuntimeConfig{
		Version:         req.Version,
		FunctionsBucket: req.FunctionsBucket,
		FunctionsPath:   req.FunctionsPath,
	}
	out, err := setup.Setup(tf, rc, req.PublicKey, req.Destroy)
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
