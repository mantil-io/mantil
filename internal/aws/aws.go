package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
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
}

func New() (*AWS, error) {
	config, err := config.LoadDefaultConfig(context.Background())
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
	}, nil
}

func (a *AWS) CreateS3Bucket(name, region string) error {
	cbi := &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(region),
		},
	}
	_, err := a.s3Client.CreateBucket(context.Background(), cbi)
	if err != nil {
		return fmt.Errorf("could not create bucket %s in %s - %v", name, region, err)
	}
	return nil
}

func (a *AWS) PutObjectToS3Bucket(bucket, key string, object io.Reader) error {
	poi := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   object,
	}

	_, err := a.s3Client.PutObject(context.Background(), poi)
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
	rsp, err := a.s3Client.GetObject(context.Background(), goi)
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

	_, err := a.s3Client.DeleteObject(context.Background(), doi)
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
		out, err := a.s3Client.ListObjectsV2(context.Background(), loi)
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

	_, err := a.s3Client.DeleteBucket(context.Background(), dbi)
	if err != nil {
		return fmt.Errorf("could not delete bucket %s - %v", name, err)
	}
	return nil
}

func (a *AWS) S3BucketExists(name string) (bool, error) {
	hbi := &s3.HeadBucketInput{
		Bucket: aws.String(name),
	}

	_, err := a.s3Client.HeadBucket(context.Background(), hbi)
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

func (a *AWS) RoleCredentials(name, role, policy string) (*stsTypes.Credentials, error) {
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
	return creds.Credentials, nil
}

func (a *AWS) UpdateLambdaFunctionCodeImage(function, image string) error {
	ufci := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(function),
		ImageUri:     aws.String(image),
	}

	_, err := a.lambdaClient.UpdateFunctionCode(context.Background(), ufci)
	if err != nil {
		return fmt.Errorf("could not update lambda function %s with image %s", function, image)
	}
	return nil
}

func (a *AWS) AccountID() (string, error) {
	gcio, err := a.stsClient.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return aws.ToString(gcio.Account), nil
}

func (a *AWS) Region() string {
	return a.config.Region
}
