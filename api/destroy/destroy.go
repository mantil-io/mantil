package destroy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/backend/api/destroy"
	"github.com/mantil-io/mantil/config"
)

type Destroy struct{}

type DestroyRequest struct {
	ProjectName string
	StageName   string
}

type DestroyResponse struct {
}

func (d *Destroy) Invoke(ctx context.Context, req *DestroyRequest) (*DestroyResponse, error) {
	return d.Destroy(ctx, req)
}

func (f *Destroy) Destroy(ctx context.Context, req *DestroyRequest) (*DestroyResponse, error) {
	if req.ProjectName == "" || req.StageName == "" {
		return nil, fmt.Errorf("bad request")
	}
	stage, err := config.LoadDeploymentState(req.ProjectName, req.StageName)
	if err != nil {
		return nil, err
	}
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	rc, err := config.LoadRuntimeConfig(awsClient)
	if err != nil {
		return nil, err
	}
	err = destroy.Destroy(req.ProjectName, stage, awsClient, rc)
	if err != nil {
		return nil, err
	}
	return &DestroyResponse{}, nil
}

func New() *Destroy {
	return &Destroy{}
}
