package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
)

func (a *AWS) APIGAtewayCloudwatchRole() (string, error) {
	gai := &apigateway.GetAccountInput{}
	gao, err := a.apigatewayClient.GetAccount(context.Background(), gai)
	if err != nil {
		return "", err
	}
	arn := aws.ToString(gao.CloudwatchRoleArn)
	if arn == "" {
		return "", nil
	}
	return iamRoleFromARN(arn)
}

func (a *AWS) SetAPIGatewayCloudwatchRole(arn string) error {
	operations := []types.PatchOperation{
		{
			Op:    types.OpReplace,
			Path:  aws.String("/cloudwatchRoleArn"),
			Value: aws.String(arn),
		},
	}
	// role could be just created and not have policy propagated returning exception that permissions are not configured
	// this request also has quota of 1 request per 20 seconds so additionally TooManyRequestException is repeated as well
	err := withRetry(func() error {
		return a.updateAPIGatewayAccount(operations)
	}, func(err error) bool {
		return strings.Contains(err.Error(), "The role ARN does not have required permissions configured.") ||
			strings.Contains(err.Error(), "API Gateway could not successfully write to CloudWatch Logs using the ARN specified") ||
			strings.Contains(err.Error(), "TooManyRequestsException")
	})
	return err
}

func (a *AWS) updateAPIGatewayAccount(operations []types.PatchOperation) error {
	uai := &apigateway.UpdateAccountInput{
		PatchOperations: operations,
	}
	_, err := a.apigatewayClient.UpdateAccount(context.Background(), uai)
	return err
}
