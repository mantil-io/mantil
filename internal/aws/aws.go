package aws

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type AWS struct {
	config   aws.Config
	s3Client *s3.Client
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
		config:   config,
		s3Client: s3.NewFromConfig(config),
	}, nil
}

func (a *AWS) DefaultRegion() string {
	return a.config.Region
}

func (a *AWS) CreateS3Bucket(name, region string) error {
	cbi := &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	}

	_, err := a.s3Client.CreateBucket(context.TODO(), cbi)
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

	_, err := a.s3Client.PutObject(context.TODO(), poi)
	if err != nil {
		return fmt.Errorf("could not put key %s in bucket %s - %v", bucket, key, err)
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
