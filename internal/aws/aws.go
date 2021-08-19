package aws

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWS struct {
	config           aws.Config
	s3Client         *s3.Client
	lambdaClient     *lambda.Client
	stsClient        *sts.Client
	ecrClient        *ecr.Client
	iamClient        *iam.Client
	cloudwatchClient *cloudwatchlogs.Client
}

func New(accessKeyID, secretAccessKey, sessionToken string) (*AWS, error) {
	config, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			sessionToken,
		)),
		config.WithRegion("eu-central-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %v", err)
	}

	if config.Region == "" {
		return nil, fmt.Errorf("default region is not specified - to specify a region either set the AWS_REGION environment variable or set the region through config file")
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
	return clientFromConfig(config), nil
}

func ListProfiles() ([]string, error) {
	configFilePath := config.DefaultSharedConfigFilename()
	buf, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not read AWS credentials file - %v", err)
	}
	profileRegex := regexp.MustCompile(`^\[profile (.*?)\]`)
	var profiles []string
	for _, line := range strings.Split(string(buf), "\n") {
		if strings.HasPrefix(line, "[default]") {
			profiles = append(profiles, "default")
			continue
		}
		res := profileRegex.FindStringSubmatch(line)
		if len(res) > 0 {
			profiles = append(profiles, res[1])
		}
	}
	return profiles, nil
}

func clientFromConfig(config aws.Config) *AWS {
	return &AWS{
		config:           config,
		s3Client:         s3.NewFromConfig(config),
		lambdaClient:     lambda.NewFromConfig(config),
		stsClient:        sts.NewFromConfig(config),
		ecrClient:        ecr.NewFromConfig(config),
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
