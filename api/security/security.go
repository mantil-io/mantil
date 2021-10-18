package security

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
)

type AWS interface {
	AccountID() string
	Region() string
	RoleCredentials(string, string, string, int32) (*aws.Credentials, error)
}

type Security struct {
	dto.SecurityRequest
	awsClient AWS
}

func New() *Security {
	return &Security{}
}

func (s *Security) Invoke(ctx context.Context, req dto.SecurityRequest) (*dto.SecurityResponse, error) {
	if err := s.init(req); err != nil {
		return nil, err
	}
	return s.credentials()
}

func (s *Security) init(req dto.SecurityRequest) error {
	awsClient, err := aws.New()
	if err != nil {
		return fmt.Errorf("error initializing aws client - %w", err)
	}
	s.SecurityRequest = req
	s.awsClient = awsClient
	return nil
}

func (s *Security) credentials() (*dto.SecurityResponse, error) {
	pptd := s.projectPolicyTemplateData()
	policy, err := s.executeProjectPolicyTemplate(pptd)
	if err != nil {
		return nil, err
	}

	creds, err := s.credentialsForPolicy(policy)
	if err != nil {
		return nil, err
	}
	return &dto.SecurityResponse{
		AccessKeyId:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		Token:           creds.SessionToken,
		Expiration:      creds.Expiration,
	}, nil
}

func (s *Security) projectPolicyTemplateData() projectPolicyTemplateData {
	pptd := projectPolicyTemplateData{
		Buckets:         s.Buckets,
		LogGroupsPrefix: s.LogGroupsPrefix,
		Region:          s.awsClient.Region(),
		AccountID:       s.awsClient.AccountID(),
	}
	return pptd
}

func (s *Security) executeProjectPolicyTemplate(pptd projectPolicyTemplateData) (string, error) {
	tpl := template.Must(template.New("").Parse(credentialsTemplate))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, pptd); err != nil {
		return "", fmt.Errorf("error executing project policy template - %w", err)
	}
	return buf.String(), nil
}

func (s *Security) credentialsForPolicy(policy string) (*aws.Credentials, error) {
	creds, err := s.awsClient.RoleCredentials("cli-user", s.CliRole, policy, 15*60)
	if err != nil {
		return nil, fmt.Errorf("error creating role credentials - %w", err)
	}
	return creds, nil
}

type projectPolicyTemplateData struct {
	Buckets         []string
	LogGroupsPrefix string
	Region          string
	AccountID       string
}
