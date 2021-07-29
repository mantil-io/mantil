package security

import (
	"context"
	"fmt"
	"log"

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
	if req.ProjectName == "" || req.Token == "" {
		return nil, fmt.Errorf("bad request")
	}

	p, err := mantil.LoadProject(req.ProjectName)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	if p.Token != req.Token {
		log.Printf("access denied - %s - %s", p.Token, req.Token)
		return nil, fmt.Errorf("access denied")
	}

	var creds *stsTypes.Credentials
	err = stream.LambdaLogStream(ctx, func() error {
		var err error
		creds, err = security.Credentials(p)
		return err
	})
	if err != nil {
		log.Printf("%v", err)
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
