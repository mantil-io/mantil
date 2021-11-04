package aws

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/endpointcreds"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/transport/http"
	"github.com/mantil-io/mantil/cli/log"
)

type AWS struct {
	config               aws.Config
	s3Client             *s3.Client
	lambdaClient         *lambda.Client
	stsClient            *sts.Client
	cloudwatchClient     *cloudwatchlogs.Client
	rgsaClient           *resourcegroupstaggingapi.Client
	dynamodbClient       *dynamodb.Client
	cloudformationClient *cloudformation.Client
	accountID            string
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
		return nil, fmt.Errorf("unable to load SDK configuration - %w", err)
	}
	if config.Region == "" {
		return nil, fmt.Errorf("aws region not set")
	}
	return clientFromConfig(config)
}

func NewWithEndpointCredentials(endpoint, region string, token func() string) (*AWS, error) {
	config, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(endpointcreds.New(endpoint,
			func(o *endpointcreds.Options) {
				o.AuthorizationToken = token()
			})),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, log.Wrap(err)
	}
	if config.Region == "" {
		return nil, log.Wrap(fmt.Errorf("aws region not set"))
	}
	return clientFromConfig(config)
}

func New() (*AWS, error) {
	config, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %w", err)
	}
	if config.Region == "" {
		return nil, fmt.Errorf("aws region not set")
	}
	return clientFromConfig(config)
}

func NewFromProfile(profile string) (*AWS, error) {
	config, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithSharedConfigProfile(profile),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %w", err)
	}
	if config.Region == "" {
		return nil, fmt.Errorf("aws region not set")
	}
	return clientFromConfig(config)
}

func clientFromConfig(config aws.Config) (*AWS, error) {
	a := &AWS{
		config:               config,
		s3Client:             s3.NewFromConfig(config),
		lambdaClient:         lambda.NewFromConfig(config),
		stsClient:            sts.NewFromConfig(config),
		cloudwatchClient:     cloudwatchlogs.NewFromConfig(config),
		rgsaClient:           resourcegroupstaggingapi.NewFromConfig(config),
		dynamodbClient:       dynamodb.NewFromConfig(config),
		cloudformationClient: cloudformation.NewFromConfig(config),
	}
	id, err := a.getAccountID()
	if err != nil {
		return nil, err
	}
	a.accountID = id
	return a, nil
}

func (a *AWS) Credentials() (aws.Credentials, error) {
	return a.config.Credentials.Retrieve(context.Background())
}

func (a *AWS) Region() string {
	return a.config.Region
}

func (a *AWS) AccountID() string {
	return a.accountID
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
		var ga *smithy.GenericAPIError
		if errors.As(err, &ga) {
			fmt.Printf("GenericAPIError: %s\n\t%s\n\t%s", ga.Code, ga.Message, ga.Fault.String())
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
	inGithubAction := os.Getenv("GITHUB_ACTIONS") == "true"
	if inGithubAction {
		cli, err := New()
		if err != nil {
			t.Fatal(err)
		}
		return cli
	}

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
