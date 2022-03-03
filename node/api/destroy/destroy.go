package destroy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/kit/aws"
	"github.com/mantil-io/mantil/node/dto"
	"github.com/mantil-io/mantil/node/terraform"
)

type Destroy struct {
	dto.DestroyRequest
	awsClient *aws.AWS
}

func New() *Destroy {
	return &Destroy{}
}

func (d *Destroy) Invoke(ctx context.Context, req dto.DestroyRequest) error {
	if err := d.init(req); err != nil {
		return err
	}

	if err := d.terraformDestroy(); err != nil {
		return fmt.Errorf("could not terraform destroy - %w", err)
	}
	if err := d.cleanupBucket(); err != nil {
		return fmt.Errorf("could not cleanup bucket - %w", err)
	}
	if err := d.cleanupResources(); err != nil {
		return fmt.Errorf("could not cleanup resources - %w", err)
	}
	return nil
}

func (d *Destroy) init(req dto.DestroyRequest) error {
	awsClient, err := aws.New()
	if err != nil {
		return fmt.Errorf("error initializing aws client - %w", err)
	}
	d.DestroyRequest = req
	d.awsClient = awsClient
	return nil
}

func (d *Destroy) terraformDestroy() error {
	tf, err := terraform.Project(d.terraformData())
	if err != nil {
		return err
	}
	return tf.Destroy()
}

func (d *Destroy) terraformData() dto.StageTemplate {
	return dto.StageTemplate{
		Project:      d.ProjectName,
		Stage:        d.StageName,
		Bucket:       d.Bucket,
		Region:       d.Region,
		BucketPrefix: d.BucketPrefix,
	}
}

func (d *Destroy) cleanupBucket() error {
	s3 := d.awsClient.S3()
	for _, prefix := range d.CleanupBucketPrefixes {
		if err := s3.DeleteBucketPrefix(d.Bucket, prefix); err != nil {
			return err
		}
	}
	return nil
}

func (d *Destroy) cleanupResources() error {
	tags := []aws.TagFilter{}
	for k, v := range d.ResourceTags {
		tags = append(tags, aws.TagFilter{Key: k, Values: []string{v}})
	}

	if err := d.awsClient.DeleteDynamodbTablesByTags(tags); err != nil {
		return err
	}
	if err := d.awsClient.S3().DeleteBucketsByTags(tags); err != nil {
		return err
	}
	return nil
}
