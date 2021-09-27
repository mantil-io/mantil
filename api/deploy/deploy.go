package deploy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/backend/api/deploy"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/terraform"
)

type Deploy struct{}

type DeployRequest struct {
	ProjectName string
	Stage       *config.Stage
}
type DeployResponse struct{}

func (h *Deploy) Init(ctx context.Context) {}

func (h *Deploy) Invoke(ctx context.Context, req *DeployRequest) (*DeployResponse, error) {
	return h.Deploy(ctx, req)
}

func (h *Deploy) Deploy(ctx context.Context, req *DeployRequest) (*DeployResponse, error) {
	if req.ProjectName == "" || req.Stage == nil {
		return nil, fmt.Errorf("bad request")
	}
	tf, err := terraform.New(req.ProjectName)
	if err != nil {
		return nil, err
	}
	defer tf.Cleanup()
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	rc, err := config.LoadRuntimeConfig(awsClient)
	if err != nil {
		return nil, err
	}
	d, err := deploy.New(req.ProjectName, req.Stage, tf, awsClient, rc)
	if err != nil {
		return nil, err
	}
	err = d.Deploy()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func New() *Deploy {
	return &Deploy{}
}
