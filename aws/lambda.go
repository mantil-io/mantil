package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

func (a *AWS) LambdaExists(name string) (bool, error) {
	gfi := &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	}
	_, err := a.lambdaClient.GetFunction(context.Background(), gfi)
	if err == nil {
		return true, nil
	}
	var rnf *lambdaTypes.ResourceNotFoundException
	if errors.As(err, &rnf) {
		return false, nil
	}
	return false, err
}

func (a *AWS) CreateLambdaFunction(name, role, s3Bucket, s3Key string, layers []string) (string, error) {
	fc := &lambdaTypes.FunctionCode{
		S3Bucket: aws.String(s3Bucket),
		S3Key:    aws.String(s3Key),
	}
	cfi := &lambda.CreateFunctionInput{
		Code:         fc,
		FunctionName: aws.String(name),
		Handler:      aws.String("bootstrap"),
		Role:         aws.String(role),
		Runtime:      lambdaTypes.RuntimeProvidedal2,
		Timeout:      aws.Int32(60 * 15),
		MemorySize:   aws.Int32(512),
		Layers:       layers,
	}
	// lambda creation might fail if the corresponding execution role was just created so we retry until it succeeds
	var rsp *lambda.CreateFunctionOutput
	var err error
	err = withRetry(func() error {
		rsp, err = a.lambdaClient.CreateFunction(context.Background(), cfi)
		return err
	}, isRetryableLambdaError)
	if err != nil {
		return "", fmt.Errorf("could not create function - %v", err)
	}
	w := lambda.NewFunctionActiveWaiter(a.lambdaClient)
	if err := w.Wait(context.Background(), &lambda.GetFunctionConfigurationInput{
		FunctionName: rsp.FunctionArn,
	}, time.Minute); err != nil {
		return "", fmt.Errorf("error waiting for function - %v", err)
	}
	return *rsp.FunctionArn, nil
}

func (a *AWS) DeleteLambdaFunction(name string) error {
	dfi := &lambda.DeleteFunctionInput{
		FunctionName: aws.String(name),
	}
	_, err := a.lambdaClient.DeleteFunction(context.Background(), dfi)
	if err != nil {
		return fmt.Errorf("error deleting lambda function - %v", err)
	}
	return nil
}

func (a *AWS) InvokeLambdaFunction(arn string, req, rsp, clientContext interface{}) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("could not marshal request - %v", err)
	}
	lii := &lambda.InvokeInput{
		FunctionName: aws.String(arn),
		Payload:      payload,
	}
	if clientContext != nil {
		buf, err := json.Marshal(clientContext)
		if err != nil {
			return fmt.Errorf("could not marshal client context - %v", err)
		}
		b64Ctx := base64.StdEncoding.EncodeToString(buf)
		lii.ClientContext = aws.String(b64Ctx)
	}
	var output *lambda.InvokeOutput
	err = withRetry(func() error {
		output, err = a.lambdaClient.Invoke(context.Background(), lii)
		return err
	}, isRetryableLambdaError)
	if err != nil {
		return fmt.Errorf("could not invoke lambda function - %v", err)
	}
	if rsp != nil && len(output.Payload) > 0 {
		if err := json.Unmarshal(output.Payload, rsp); err != nil {
			return fmt.Errorf("could not unmarshal response - %v", err)
		}
	}
	return nil
}

func (a *AWS) UpdateLambdaFunctionCodeFromS3(function, bucket, key string) error {
	ufci := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(function),
		S3Bucket:     aws.String(bucket),
		S3Key:        aws.String(key),
	}

	_, err := a.lambdaClient.UpdateFunctionCode(context.Background(), ufci)
	if err != nil {
		return fmt.Errorf("could not update lambda function %s from %s/%s - %v", function, bucket, key, err)
	}
	return nil
}

func (a *AWS) WaitLambdaFunctionUpdated(function string) error {
	gfci := &lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(function),
	}

	retryInterval := 5 * time.Second
	retryAttempts := 60
	for retryAttempts > 0 {
		gfco, err := a.lambdaClient.GetFunctionConfiguration(context.Background(), gfci)
		if err != nil {
			return err
		}
		if gfco.LastUpdateStatus == lambdaTypes.LastUpdateStatusSuccessful {
			return nil
		}
		if gfco.LastUpdateStatus == lambdaTypes.LastUpdateStatusFailed {
			return errors.New(*gfco.LastUpdateStatusReason)
		}
		time.Sleep(retryInterval)
		retryAttempts--
	}
	return nil
}

func isRetryableLambdaError(err error) bool {
	// lambda tends to return generic error codes (eg. ErrCodeInvalidParameterValueException)
	// so we resort to matching error message strings here. Reference:
	// https://github.com/hashicorp/terraform-provider-aws/blob/ac06ced75cba0daf09fef2938752ad13cc6fff6e/aws/resource_aws_lambda_function.go#L526
	return strings.Contains(err.Error(), "The role defined for the function cannot be assumed by Lambda") ||
		strings.Contains(err.Error(), "The provided execution role does not have permissions")
}
