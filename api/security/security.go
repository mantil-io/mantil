package security

import (
	"context"
	"fmt"

	"github.com/atoz-technology/mantil-backend/internal/mantil"
	"github.com/atoz-technology/mantil-backend/internal/security"
)

type Security struct{}

type SecurityRequest struct {
	ProjectName string
	Token       string
}

type SecurityResponse struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func (f *Security) Invoke(ctx context.Context, req *SecurityRequest) (*SecurityResponse, error) {
	return f.FederationToken(ctx, req)
}

func (f *Security) FederationToken(ctx context.Context, req *SecurityRequest) (*SecurityResponse, error) {
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

	creds, err := security.FederationToken(p)
	if err != nil {
		return nil, err
	}

	return &SecurityResponse{
		AccessKeyID:     *creds.AccessKeyId,
		SecretAccessKey: *creds.SecretAccessKey,
		SessionToken:    *creds.SessionToken,
	}, nil
}

func New() *Security {
	return &Security{}
}
