package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWS struct {
	config       aws.Config
	s3Client     *s3.Client
	lambdaClient *lambda.Client
	stsClient    *sts.Client
	ecrClient    *ecr.Client
}

func New(accessKeyID, secretAccessKey, sessionToken string) (*AWS, error) {
	config, err := config.LoadDefaultConfig(context.TODO(),
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
		context.TODO(),
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
		config:       config,
		s3Client:     s3.NewFromConfig(config),
		lambdaClient: lambda.NewFromConfig(config),
		stsClient:    sts.NewFromConfig(config),
		ecrClient:    ecr.NewFromConfig(config),
	}
}

func (a *AWS) Credentials() (aws.Credentials, error) {
	return a.config.Credentials.Retrieve(context.TODO())
}

func (a *AWS) PutObjectToS3Bucket(bucket, key string, object io.Reader) error {
	poi := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   object,
	}

	_, err := a.s3Client.PutObject(context.TODO(), poi)
	if err != nil {
		return fmt.Errorf("could not put key %s in bucket %s - %v", bucket, key, err)
	}
	return nil
}

func (a *AWS) GetObjectFromS3Bucket(bucket, key string, o interface{}) error {
	goi := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	rsp, err := a.s3Client.GetObject(context.TODO(), goi)
	if err != nil {
		return fmt.Errorf("could not get key %s from bucket %s - %v", bucket, key, err)
	}
	defer rsp.Body.Close()

	decoder := json.NewDecoder(rsp.Body)
	if err := decoder.Decode(&o); err != nil {
		return err
	}
	return nil
}

func (a *AWS) GetECRLogin() (string, string, error) {
	geto, err := a.ecrClient.GetAuthorizationToken(context.TODO(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return "", "", err
	}
	if len(geto.AuthorizationData) == 0 || geto.AuthorizationData[0].AuthorizationToken == nil {
		return "", "", fmt.Errorf("no authorization data returned for ECR")
	}

	at := *geto.AuthorizationData[0].AuthorizationToken
	dat, err := base64.StdEncoding.DecodeString(at)
	if err != nil {
		return "", "", err
	}

	login := strings.Split(string(dat), ":")
	if len(login) != 2 {
		return "", "", fmt.Errorf("login data wrong format")
	}
	return login[0], login[1], nil
}
