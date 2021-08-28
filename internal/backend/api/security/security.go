package security

import (
	"bytes"
	"fmt"
	"html/template"

	stsTypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/mantil"
)

func Credentials(project *mantil.Project) (*stsTypes.Credentials, error) {
	aws, err := aws.New()
	if err != nil {
		return nil, err
	}
	accountID, err := aws.AccountID()
	if err != nil {
		return nil, err
	}
	policy, err := fillProjectPolicyTemplate(project, accountID)
	if err != nil {
		return nil, err
	}
	role := fmt.Sprintf("arn:aws:iam::%s:role/mantil-cli-user", accountID)
	creds, err := aws.RoleCredentials(project.Name, role, policy)
	if err != nil {
		return nil, err
	}
	return creds, nil
}

func fillProjectPolicyTemplate(project *mantil.Project, accountID string) (string, error) {
	ppt := ProjectPolicyTemplate{
		Name:             project.Name,
		OrganizationName: project.Organization.Name,
		Bucket:           project.Bucket,
		Region:           "eu-central-1",
		AccountID:        accountID,
		StaticWebsites:   project.StaticWebsites,
	}
	tpl := template.Must(template.New("").Parse(CredentialsTemplate))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, ppt); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type ProjectPolicyTemplate struct {
	Name             string
	OrganizationName string
	Bucket           string
	Region           string
	AccountID        string
	StaticWebsites   []mantil.StaticWebsite
}
