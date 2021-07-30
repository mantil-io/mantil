package security

import (
	"context"
	"fmt"

	"github.com/atoz-technology/mantil-backend/internal/mantil"
	"github.com/atoz-technology/mantil-backend/internal/security"
	"github.com/atoz-technology/mantil-backend/internal/stream"
	stsTypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
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
	creds, err := f.streamingLogsCredentials(ctx, p)
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
	return req.ProjectName != "" && req.Token != ""
}

func (f *Security) streamingLogsCredentials(ctx context.Context, p *mantil.Project) (*stsTypes.Credentials, error) {
	var creds *stsTypes.Credentials
	err := stream.LambdaLogStream(ctx, func() error {
		var err error
		creds, err = security.Credentials(p)
		return err
	})
	return creds, err
}

func New() *Security {
	return &Security{}
}
