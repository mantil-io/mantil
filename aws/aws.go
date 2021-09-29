package aws

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/transport/http"
)

type AWS struct {
	config           aws.Config
	s3Client         *s3.Client
	lambdaClient     *lambda.Client
	stsClient        *sts.Client
	iamClient        *iam.Client
	cloudwatchClient *cloudwatchlogs.Client
}

func NewWithCredentials(accessKeyID, secretAccessKey, sessionToken, region string) (*AWS, error) {
	config, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			sessionToken)),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %v", err)
	}
	if config.Region == "" {
		return nil, fmt.Errorf("aws region not set")
	}
	return clientFromConfig(config), nil
}

func New() (*AWS, error) {
	config, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %v", err)
	}
	if config.Region == "" {
		return nil, fmt.Errorf("aws region not set")
	}
	return clientFromConfig(config), nil
}

func NewFromProfile(profile string) (*AWS, error) {
	config, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %v", err)
	}
	if config.Region == "" {
		return nil, fmt.Errorf("aws region not set")
	}
	return clientFromConfig(config), nil
}

func clientFromConfig(config aws.Config) *AWS {
	return &AWS{
		config:           config,
		s3Client:         s3.NewFromConfig(config),
		lambdaClient:     lambda.NewFromConfig(config),
		stsClient:        sts.NewFromConfig(config),
		iamClient:        iam.NewFromConfig(config),
		cloudwatchClient: cloudwatchlogs.NewFromConfig(config),
	}
}

func (a *AWS) Credentials() (aws.Credentials, error) {
	return a.config.Credentials.Retrieve(context.Background())
}

func (a *AWS) Region() string {
	return a.config.Region
}

// examineError helper for finding type of the aws error which can be use in error.As
func examineError(err error) {
	var oe *smithy.OperationError
	if errors.As(err, &oe) {
		fmt.Printf("OperationError: failed to call service: %s\n\toperation: %s\n\terror: %s\n", oe.Service(), oe.Operation(), oe.Unwrap())
		fmt.Printf("\tinner error type: %T\n", oe.Unwrap())
		var he *http.ResponseError
		if errors.As(err, &he) {
			fmt.Printf("httpRespose status code: %d\n\terror: %s\n", he.HTTPStatusCode(), he.Unwrap())
			fmt.Printf("\tinner error type  %T\n", he.Unwrap())
		}
	}
}

const testsProfileEnv = "MANTIL_TESTS_AWS_PROFILE"

type testingI interface {
	Logf(format string, args ...interface{})
	Fatal(args ...interface{})
}

// NewForTests creates new client for use in tests
// environment variable with aws profile must be declare
// testingI is here to brake dependency on testing package
// Usage:
//  func TestMy(t *testing.T) {
//  	cli := NewForTests(t)
//  	if cli == nil {
// 			t.Skip("skip: cli not initialized")
//   	}
//
// Run tests:
//  MANTIL_TESTS_AWS_PROFILE=org5 go test -v
func NewForTests(t testingI) *AWS {
	val, ok := os.LookupEnv(testsProfileEnv)
	if !ok {
		t.Logf("environment vairable %s not found", testsProfileEnv)
		return nil
	}
	cli, err := NewFromProfile(val)
	if err != nil {
		t.Fatal(err)
	}
	return cli

}

func TestProfile() string {
	val, ok := os.LookupEnv(testsProfileEnv)
	if !ok {
		return ""
	}
	return val
}
