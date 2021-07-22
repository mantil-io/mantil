package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	stsTypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/aws/smithy-go"
)

type AWS struct {
	config       aws.Config
	s3Client     *s3.Client
	lambdaClient *lambda.Client
	stsClient    *sts.Client
	ecrClient    *ecr.Client
}

func New() (*AWS, error) {
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %v", err)
	}

	if config.Region == "" {
		return nil, fmt.Errorf("default region is not specified - to specify a region either set the AWS_REGION environment variable or set the region through config file")
	}

	return &AWS{
		config:       config,
		s3Client:     s3.NewFromConfig(config),
		lambdaClient: lambda.NewFromConfig(config),
		stsClient:    sts.NewFromConfig(config),
		ecrClient:    ecr.NewFromConfig(config),
	}, nil
}

func (a *AWS) Credentials() (aws.Credentials, error) {
	return a.config.Credentials.Retrieve(context.TODO())
}

func (a *AWS) DefaultRegion() string {
	return a.config.Region
}

func (a *AWS) CreateS3Bucket(name, region, projectTag string) error {
	cbi := &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(region),
		},
	}
	_, err := a.s3Client.CreateBucket(context.TODO(), cbi)
	if err != nil {
		return fmt.Errorf("could not create bucket %s in %s - %v", name, region, err)
	}
	bti := &s3.PutBucketTaggingInput{
		Bucket: aws.String(name),
		Tagging: &s3Types.Tagging{
			TagSet: []s3Types.Tag{
				{Key: aws.String("access-project"), Value: aws.String(projectTag)},
			},
		},
	}
	_, err = a.s3Client.PutBucketTagging(context.TODO(), bti)
	if err != nil {
		return fmt.Errorf("could not set bucket tagging - %v", err)
	}
	return nil
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

func (a *AWS) DeleteObjectFromS3Bucket(bucket, key string) error {
	doi := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := a.s3Client.DeleteObject(context.TODO(), doi)
	if err != nil {
		return fmt.Errorf("could not delete key %s in bucket %s - %v", bucket, key, err)
	}
	return nil
}

func (a *AWS) EmptyS3Bucket(name string) error {
	loi := &s3.ListObjectsV2Input{
		Bucket: aws.String(name),
	}

	for {
		out, err := a.s3Client.ListObjectsV2(context.TODO(), loi)
		if err != nil {
			return fmt.Errorf("emptying bucket failed - %v", err)
		}

		for _, object := range out.Contents {
			if err := a.DeleteObjectFromS3Bucket(name, *object.Key); err != nil {
				return fmt.Errorf("emptying bucket failed - %v", err)
			}
		}

		if out.IsTruncated {
			loi.ContinuationToken = out.ContinuationToken
		} else {
			break
		}
	}
	return nil
}

func (a *AWS) DeleteS3Bucket(name string) error {
	// all objects must be deleted before bucket can be deleted
	if err := a.EmptyS3Bucket(name); err != nil {
		return fmt.Errorf("could not delete bucket %s - %v", name, err)
	}

	dbi := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}

	_, err := a.s3Client.DeleteBucket(context.TODO(), dbi)
	if err != nil {
		return fmt.Errorf("could not delete bucket %s - %v", name, err)
	}
	return nil
}

func (a *AWS) S3BucketExists(name string) (bool, error) {
	hbi := &s3.HeadBucketInput{
		Bucket: aws.String(name),
	}

	_, err := a.s3Client.HeadBucket(context.TODO(), hbi)
	if err != nil {
		var oe smithy.APIError
		if errors.As(err, &oe) {
			switch oe.ErrorCode() {
			case "Forbidden":
				return true, nil
			case "NotFound":
				return false, nil
			case "MovedPermanently":
				return true, nil
			default:
				return false, err
			}
		} else {
			return false, err
		}
	}
	return true, nil
}

func (a *AWS) UpdateLambdaFunctionCodeFromS3(function, bucket, key string) error {
	ufci := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(function),
		S3Bucket:     aws.String(bucket),
		S3Key:        aws.String(key),
	}

	_, err := a.lambdaClient.UpdateFunctionCode(context.TODO(), ufci)
	if err != nil {
		return fmt.Errorf("could not update lambda function %s from %s/%s - %v", function, bucket, key, err)
	}
	return nil
}

func (a *AWS) GetProjectToken(name, policy string) (*stsTypes.Credentials, error) {
	gfti := &sts.GetFederationTokenInput{
		DurationSeconds: aws.Int32(900),
		Name:            aws.String(name),
		Policy:          aws.String(policy),
	}

	rsp, err := a.stsClient.GetFederationToken(context.TODO(), gfti)
	if err != nil {
		return nil, fmt.Errorf("could not get project token - %v", err)
	}
	return rsp.Credentials, nil
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
