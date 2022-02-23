package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func (a *AWS) GetSSMParameter(path string) (string, error) {
	o, err := a.ssmClient.GetParameter(context.Background(), &ssm.GetParameterInput{
		Name:           aws.String(path),
		WithDecryption: true,
	})
	if err != nil {
		return "", err
	}
	return *o.Parameter.Value, nil
}
