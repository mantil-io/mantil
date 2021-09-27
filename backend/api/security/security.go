package security

import (
	"bytes"
	"fmt"
	"html/template"

	stsTypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/config"
)

func Credentials(projectName string, stage *config.Stage) (*stsTypes.Credentials, string, error) {
	aws, err := aws.New()
	if err != nil {
		return nil, "", err
	}
	accountID, err := aws.AccountID()
	if err != nil {
		return nil, "", err
	}
	policy, err := fillProjectPolicyTemplate(projectName, stage, accountID, aws)
	if err != nil {
		return nil, "", err
	}
	role := fmt.Sprintf("arn:aws:iam::%s:role/mantil-cli-user", accountID)
	creds, err := aws.RoleCredentials(projectName, role, policy)
	if err != nil {
		return nil, "", err
	}
	return creds, aws.Region(), nil
}

func fillProjectPolicyTemplate(projectName string, stage *config.Stage, accountID string, aws *aws.AWS) (string, error) {
	bucket, err := config.Bucket(aws)
	if err != nil {
		return "", err
	}
	ppt := ProjectPolicyTemplate{
		Name:      projectName,
		Bucket:    bucket,
		Region:    aws.Region(),
		AccountID: accountID,
	}
	if stage != nil {
		ppt.PublicSites = stage.PublicSites
		ppt.LogGroup = config.ProjectResource(projectName, stage.Name)
	}
	tpl := template.Must(template.New("").Parse(CredentialsTemplate))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, ppt); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type ProjectPolicyTemplate struct {
	Name        string
	Bucket      string
	Region      string
	AccountID   string
	PublicSites []*config.PublicSite
	LogGroup    string
}
