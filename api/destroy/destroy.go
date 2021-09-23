package destroy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/internal/backend/api/destroy"
	"github.com/mantil-io/mantil/internal/mantil"
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
	project, err := mantil.LoadProjectS3(req.ProjectName)
	if err != nil {
		return nil, err
	}
	err = destroy.Destroy(project, req.Stage)
	if err != nil {
		return nil, err
	}
	return &DestroyResponse{}, nil
}

func New() *Destroy {
	return &Destroy{}
}
