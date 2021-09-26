package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	stsTypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
)

func (a *AWS) AccountID() (string, error) {
	gcio, err := a.stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("could not get account ID - %w", err)
	}
	return aws.ToString(gcio.Account), nil
}

func (a *AWS) RoleCredentials(name, role, policy string) (*stsTypes.Credentials, error) {
	ari := &sts.AssumeRoleInput{
		RoleArn:         aws.String(role),
		RoleSessionName: aws.String(name),
		DurationSeconds: aws.Int32(900),
		Policy:          aws.String(policy),
	}

	creds, err := a.stsClient.AssumeRole(context.Background(), ari)
	if err != nil {
		return nil, err
	}
	return creds.Credentials, nil
}
