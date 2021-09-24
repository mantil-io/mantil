package deploy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/backend/api/deploy"
	"github.com/mantil-io/mantil/internal/backend/terraform"
	"github.com/mantil-io/mantil/internal/mantil"
)

type Deploy struct{}

type DeployRequest struct {
	ProjectName string
	Stage       *mantil.Stage
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
	project, err := mantil.LoadProjectS3(req.ProjectName)
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
	rc, err := mantil.LoadRuntimeConfig(awsClient)
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
