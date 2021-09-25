package setup

import (
	"context"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/backend/api/setup"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/terraform"
)

type Setup struct{}

func (f *Setup) Invoke(ctx context.Context, req *dto.SetupRequest) (*dto.SetupResponse, error) {
	return f.Setup(ctx, req)
}

func (f *Setup) Setup(ctx context.Context, req *dto.SetupRequest) (*dto.SetupResponse, error) {
	tf, err := terraform.New("mantil-setup")
	if err != nil {
		return nil, err
	}
	defer tf.Cleanup()
	rc := &config.RuntimeConfig{
		// TODO: sto ce mi ovaj version kada se nigdje ne koristi
		Version:         req.Version,
		FunctionsBucket: req.FunctionsBucket,
		FunctionsPath:   req.FunctionsPath,
	}
	out, err := setup.Setup(tf, rc, req.PublicKey, req.Destroy)
	if err != nil {
		return nil, err
	}
	return &dto.SetupResponse{
		APIGatewayRestURL: out.RestURL,
		APIGatewayWsURL:   out.WsURL,
	}, nil
}

func New() *Setup {
	return &Setup{}
}
