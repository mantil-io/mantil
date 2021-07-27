package destroy

import (
	"context"
	"fmt"
	"log"

	"github.com/atoz-technology/mantil-backend/internal/destroy"
	"github.com/atoz-technology/mantil-backend/internal/mantil"
)

type Destroy struct{}

type DestroyRequest struct {
	ProjectName string
	Token       string
}

type DestroyResponse struct {
}

func (d *Destroy) Invoke(ctx context.Context, req *DestroyRequest) (*DestroyResponse, error) {
	return d.Destroy(ctx, req)
}

func (f *Destroy) Destroy(ctx context.Context, req *DestroyRequest) (*DestroyResponse, error) {
	if req.ProjectName == "" || req.Token == "" {
		return nil, fmt.Errorf("bad request")
	}
	p, err := mantil.LoadProject(req.ProjectName)
	if err != nil {
		return nil, err
	}
	if p.Token != req.Token {
		return nil, fmt.Errorf("access denied")
	}
	err = destroy.Destroy(p, "/tmp")
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	return &DestroyResponse{}, nil
}

func New() *Destroy {
	return &Destroy{}
}
