package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type SDK struct {
	config aws.Config
}

func NewSDK() (*SDK, error) {
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK configuration - %v", err)
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

	lc := s.BucketLocationConstraintForRegion(region)
	if lc == "" {
		return fmt.Errorf("region is not one of the known values - %s", region)
	}

	cbi := &s3.CreateBucketInput{
		Bucket: aws.String(name),
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: lc,
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

func (s *SDK) BucketLocationConstraintForRegion(region string) types.BucketLocationConstraint {
	for _, blc := range types.BucketLocationConstraint.Values() {
		if blc == region {
			return blc
		}
	}
	return ""
}
