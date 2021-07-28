package security

import (
	"bytes"
	"html/template"

	"github.com/atoz-technology/mantil-backend/internal/aws"
	"github.com/atoz-technology/mantil-backend/internal/mantil"
	stsTypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
)

func FederationToken(project *mantil.Project) (*stsTypes.Credentials, error) {
	aws, err := aws.New()
	if err != nil {
		return nil, err
	}
	accountID, err := aws.AccountID()
	if err != nil {
		return nil, err
	}

	ppt := ProjectPolicyTemplate{
		Name:             project.Name,
		OrganizationName: project.Organization.Name,
		Bucket:           project.Bucket,
		Region:           "eu-central-1",
		AccountID:        accountID,
	}

	tpl := template.Must(template.New("").Parse(FederationTokenPolicyTemplate))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, ppt); err != nil {
		return nil, err
	}

	creds, err := aws.GetProjectToken(project.Name, buf.String())
	if err != nil {
		return nil, err
	}
	return creds, nil
}

type ProjectPolicyTemplate struct {
	Name             string
	OrganizationName string
	Bucket           string
	Region           string
	AccountID        string
}
