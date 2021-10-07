package security

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/mantil-io/mantil/aws"

	"github.com/mantil-io/mantil/workspace"
)

type AWS interface {
	AccountID() string
	Region() string
	RoleCredentials(string, string, string) (*aws.Credentials, error)
}

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

type Security struct {
	req        *SecurityRequest
	stage      *workspace.Stage
	bucketName string
	awsClient  AWS
}

func New() *Security {
	return &Security{}
}

func (s *Security) Invoke(ctx context.Context, req *SecurityRequest) (*SecurityResponse, error) {
	if err := s.init(req); err != nil {
		return nil, err
	}
	return s.credentials()
}

func (s *Security) init(req *SecurityRequest) error {
	awsClient, err := aws.New()
	if err != nil {
		return fmt.Errorf("error initializing aws client - %w", err)
	}
	var stage *workspace.Stage
	if req.StageName != "" {
		// ignore this error as deployment state won't exist for newly created stages
		stage, _ = workspace.LoadStageState(req.ProjectName, req.StageName)
	}
	bucketName, err := workspace.Bucket(awsClient)
	if err != nil {
		return err
	}
	s.req = req
	s.stage = stage
	s.bucketName = bucketName
	s.awsClient = awsClient
	return nil
}

func (s *Security) credentials() (*SecurityResponse, error) {
	ppt, err := s.projectPolicyTemplateData()
	if err != nil {
		return nil, err
	}
	policy, err := s.executeProjectPolicyTemplate(ppt)
	if err != nil {
		return nil, err
	}
	creds, err := s.credentialsForPolicy(policy)
	if err != nil {
		return nil, err
	}
	return &SecurityResponse{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Region:          creds.Region,
	}, nil
}

func (s *Security) projectPolicyTemplateData() (*projectPolicyTemplateData, error) {
	ppt := &projectPolicyTemplateData{
		Name:      s.req.ProjectName,
		Bucket:    s.bucketName,
		Region:    s.awsClient.Region(),
		AccountID: s.awsClient.AccountID(),
	}
	if s.stage != nil {
		ppt.Public = s.stage.Public
		ppt.LogGroup = workspace.ProjectResource(s.req.ProjectName, s.stage.Name)
	}
	return ppt, nil
}

func (s *Security) executeProjectPolicyTemplate(pptd *projectPolicyTemplateData) (string, error) {
	tpl := template.Must(template.New("").Parse(CredentialsTemplate))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, pptd); err != nil {
		return "", fmt.Errorf("error executing project policy template - %w", err)
	}
	return buf.String(), nil
}

func (s *Security) credentialsForPolicy(policy string) (*credentials, error) {
	role := s.cliUserRole()
	creds, err := s.awsClient.RoleCredentials(s.req.ProjectName, role, policy)
	if err != nil {
		return nil, fmt.Errorf("error creating role credentials - %w", err)
	}
	return &credentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Region:          s.awsClient.Region(),
	}, nil
}

func (s *Security) cliUserRole() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/mantil-cli-user", s.awsClient.AccountID())
}

type projectPolicyTemplateData struct {
	Name      string
	Bucket    string
	Region    string
	AccountID string
	Public    []*workspace.PublicSite
	LogGroup  string
}

type credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}
