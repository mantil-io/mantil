package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

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
	retryInterval := time.Second
	retryAttempts := 60
	var rsp *lambda.CreateFunctionOutput
	var err error
	for retryAttempts > 0 {
		rsp, err = a.lambdaClient.CreateFunction(context.TODO(), cfi)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "The role defined for the function cannot be assumed by Lambda") ||
			strings.Contains(err.Error(), "The provided execution role does not have permissions") {
			time.Sleep(retryInterval)
			retryAttempts--
			continue
		}
		if err != nil {
			return "", fmt.Errorf("could not create function - %v", err)
		}
	}
	w := lambda.NewFunctionActiveWaiter(a.lambdaClient)
	if err := w.Wait(context.TODO(), &lambda.GetFunctionConfigurationInput{
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
	_, err := a.lambdaClient.DeleteFunction(context.TODO(), dfi)
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
	output, err := a.lambdaClient.Invoke(context.TODO(), lii)
	if err != nil {
		return fmt.Errorf("could not invoke lambda function - %v", err)
	}
	if rsp != nil {
		if err := json.Unmarshal(output.Payload, rsp); err != nil {
			return fmt.Errorf("could not unmarshal response - %v", err)
		}
	}
	return nil
}
