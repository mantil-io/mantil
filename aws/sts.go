package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

func (a *AWS) getAccountID() (string, error) {
	gcio, err := a.stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("could not get account ID - %w", err)
	}
	return aws.ToString(gcio.Account), nil
}

func (a *AWS) RoleCredentials(name, role, policy string) (*Credentials, error) {
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
	return &Credentials{
		AccessKeyID:     aws.ToString(creds.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(creds.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(creds.Credentials.SessionToken),
	}, nil
}

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// Try will return error if can't get AWS account id
func (a *AWS) Try() error {
	_, err := a.stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		//examineError(err)
		return asCredentialsError(err)
	}
	return nil
}

func asCredentialsError(err error) error {
	var ga *smithy.GenericAPIError
	if errors.As(err, &ga) {
		if ga.Code == "InvalidClientTokenId" {
			return fmt.Errorf("invalid AWS access credentials")
		}
	}
	var me *retry.MaxAttemptsError
	if errors.As(err, &me) {
		return fmt.Errorf("check your credentials, exceeded maximum number of attempts %d", me.Attempt)
	}

	return err
}
