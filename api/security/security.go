package security

import (
	"context"
	"fmt"
	"github.com/mantil-io/mantil/backend/api/security"
	"github.com/mantil-io/mantil/config"
)

type Security struct{}

type SecurityRequest struct {
	ProjectName string
	StageName   string
}

type SecurityResponse struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}

func (f *Security) Invoke(ctx context.Context, req *SecurityRequest) (*SecurityResponse, error) {
	return f.Credentials(ctx, req)
}

func (f *Security) Credentials(ctx context.Context, req *SecurityRequest) (*SecurityResponse, error) {
	if !f.isRequestValid(req) {
		return nil, fmt.Errorf("bad request")
	}
	project, err := config.LoadProjectS3(req.ProjectName)
	if err != nil {
		return nil, err
	}
	stage := project.Stage(req.StageName)
	creds, region, err := security.Credentials(project, stage)
	if err != nil {
		return nil, err
	}
	return &SecurityResponse{
		AccessKeyID:     *creds.AccessKeyId,
		SecretAccessKey: *creds.SecretAccessKey,
		SessionToken:    *creds.SessionToken,
		Region:          region,
	}, nil
}

func (f *Security) isRequestValid(req *SecurityRequest) bool {
	if req == nil {
		return false
	}
	return req.ProjectName != ""
}

func New() *Security {
	return &Security{}
}
