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
	if req.ProjectName == "" {
		return nil, fmt.Errorf("bad request")
	}
	project, err := config.LoadProjectS3(req.ProjectName)
	if err != nil {
		return nil, err
	}
	tf, err := terraform.New(project.Name)
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
	d, err := deploy.New(project, req.Stage, tf, awsClient, rc)
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
