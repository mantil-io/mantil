package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func (a *AWS) AccountID() (string, error) {
	gcio, err := a.stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("could not get account ID - %v", err)
	}
	return aws.ToString(gcio.Account), nil
}
