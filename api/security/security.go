package security

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/internal/backend/api/security"
	"github.com/mantil-io/mantil/internal/mantil"
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
	return f.Credentials(ctx, req)
}

func (f *Security) Credentials(ctx context.Context, req *SecurityRequest) (*SecurityResponse, error) {
	if !f.isRequestValid(req) {
		return nil, fmt.Errorf("bad request")
	}
	p, err := mantil.LoadProject(req.ProjectName)
	if err != nil {
		return nil, err
	}
	if !p.IsValidToken(req.Token) {
		return nil, fmt.Errorf("access denied")
	}
	creds, err := security.Credentials(p)
	if err != nil {
		return nil, err
	}
	return &SecurityResponse{
		AccessKeyID:     *creds.AccessKeyId,
		SecretAccessKey: *creds.SecretAccessKey,
		SessionToken:    *creds.SessionToken,
	}, nil
}

func (f *Security) isRequestValid(req *SecurityRequest) bool {
	if req == nil {
		return false
	}
	return req.ProjectName != "" && req.Token != ""
}

func New() *Security {
	return &Security{}
}
