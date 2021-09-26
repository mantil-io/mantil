package security

import (
	"bytes"
	"context"
	"fmt"
	"html/template"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/config"
)

type Security struct {
	awsClient *aws.AWS
	project   *config.Project
	stage     *config.Stage
}

func New() *Security {
	return &Security{}
}

func (s *Security) Invoke(ctx context.Context, req *dto.SecurityRequest) (*dto.SecurityResponse, error) {
	if err := s.init(req); err != nil {
		return nil, err
	}
	return s.credentials()
}

func (s *Security) init(req *dto.SecurityRequest) error {
	awsClient, err := aws.New()
	if err != nil {
		return fmt.Errorf("error initializing aws client - %w", err)
	}
	project, err := config.LoadProjectS3(req.ProjectName)
	if err != nil {
		return fmt.Errorf("error fetching project %s - %w", req.ProjectName, err)
	}
	s.awsClient = awsClient
	s.project = project
	s.stage = project.Stage(req.StageName)
	return nil
}

func (s *Security) credentials() (*dto.SecurityResponse, error) {
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
	return &dto.SecurityResponse{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
		Region:          creds.Region,
	}, nil
}

func (s *Security) projectPolicyTemplateData() (*projectPolicyTemplateData, error) {
	accountID, err := s.awsClient.AccountID()
	if err != nil {
		return nil, fmt.Errorf("error fetching aws account id - %w", err)
	}
	ppt := &projectPolicyTemplateData{
		Name:        s.project.Name,
		Bucket:      s.project.Bucket,
		Region:      s.awsClient.Region(),
		AccountID:   accountID,
		PublicSites: s.stage.PublicSites,
		LogGroup:    config.ProjectResource(s.project.Name, s.stage.Name),
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
	role, err := s.cliUserRole()
	if err != nil {
		return nil, err
	}
	creds, err := s.awsClient.RoleCredentials(s.project.Name, role, policy)
	if err != nil {
		return nil, fmt.Errorf("error creating role credentials - %w", err)
	}
	return &credentials{
		AccessKeyID:     *creds.AccessKeyId,
		SecretAccessKey: *creds.SecretAccessKey,
		SessionToken:    *creds.SessionToken,
		Region:          s.awsClient.Region(),
	}, nil
}

func (s *Security) cliUserRole() (string, error) {
	accountID, err := s.awsClient.AccountID()
	if err != nil {
		return "", fmt.Errorf("error fetching aws account id - %w", err)
	}
	return fmt.Sprintf("arn:aws:iam:%s:role/mantil-cli-user", accountID), nil
}

type projectPolicyTemplateData struct {
	Name        string
	Bucket      string
	Region      string
	AccountID   string
	PublicSites []*config.PublicSite
	LogGroup    string
}

type credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}
