package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type AWS struct {
	config aws.Config
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
		config: config,
	}, nil
}

func (a *AWS) DefaultRegion() string {
	return a.config.Region
}

func (a *AWS) CreateS3Bucket(name, region string) error {
	client := s3.NewFromConfig(a.config)

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

func (a *AWS) DeleteS3Bucket(name string) error {
	client := s3.NewFromConfig(a.config)

	dbi := &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	}

	_, err := client.DeleteBucket(context.TODO(), dbi)
	if err != nil {
		return fmt.Errorf("could not delete bucket %s - %v", name, err)
	}
	return nil
}

func (a *AWS) S3BucketExists(name string) (bool, error) {
	client := s3.NewFromConfig(a.config)

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

// TODO: check what type of object in arguments is most convenient
func (a *AWS) PutObjectToS3Bucket(bucket, key string, object *bytes.Reader) error {
	client := s3.NewFromConfig(a.config)

	poi := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   object,
	}

	_, err := client.PutObject(context.TODO(), poi)
	if err != nil {
		return fmt.Errorf("could not put key %s in bucket %s - %v", bucket, key, err)
	}
	return nil
}
