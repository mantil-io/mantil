package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type SDK struct {
	config aws.Config
}

func NewSDK() (*SDK, error) {
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %v", err)
	}

	if config.Region == "" {
		return nil, fmt.Errorf("default region is not specified - to specify a region either set the AWS_REGION environment variable or set the region through config file")
	}

	return &SDK{
		config: config,
	}, nil
}

func (s *SDK) DefaultRegion() string {
	return s.config.Region
}

func (s *SDK) CreateS3Bucket(name, region string) error {
	client := s3.NewFromConfig(s.config)

	cbi := &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	}

	_, err := client.CreateBucket(context.TODO(), cbi)
	if err != nil {
		return fmt.Errorf("could not create bucket %s in %s - %v", name, region, err)
	}
	return nil
}

func (s *SDK) DeleteS3Bucket(name string) error {
	client := s3.NewFromConfig(s.config)

	dbi := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}

	_, err := client.DeleteBucket(context.TODO(), dbi)
	if err != nil {
		return fmt.Errorf("could not delete bucket %s - %v", name, err)
	}
	return nil
}

func (s *SDK) S3BucketExists(name string) (bool, error) {
	client := s3.NewFromConfig(s.config)

	hbi := &s3.HeadBucketInput{
		Bucket: aws.String(name),
	}

	_, err := client.HeadBucket(context.TODO(), hbi)
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
