package destroy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/backend/api/destroy"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/terraform"
)

type Destroy struct{}

type DestroyRequest struct {
	ProjectName string
	Stage       string
}

type DestroyResponse struct {
}

func (d *Destroy) Invoke(ctx context.Context, req *DestroyRequest) (*DestroyResponse, error) {
	return d.Destroy(ctx, req)
}

func (f *Destroy) Destroy(ctx context.Context, req *DestroyRequest) (*DestroyResponse, error) {
	if req.ProjectName == "" {
		return nil, fmt.Errorf("bad request")
	}
	project, err := config.LoadProjectS3(req.ProjectName)
	if err != nil {
		return nil, err
	}
	tf, err := terraform.New(fmt.Sprintf("%s-%s", project.Name, req.Stage))
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
	err = destroy.Destroy(project, req.Stage, tf, awsClient, rc)
	if err != nil {
		return nil, err
	}
	return &DestroyResponse{}, nil
}

func New() *Destroy {
	return &Destroy{}
}
