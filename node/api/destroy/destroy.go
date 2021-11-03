package destroy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/node/dto"
	"github.com/mantil-io/mantil/node/terraform"
)

type Destroy struct {
	dto.DestroyRequest
}

func New() *Destroy {
	return &Destroy{}
}

func (d *Destroy) Invoke(ctx context.Context, req dto.DestroyRequest) error {
	d.DestroyRequest = req

	if err := d.terraformDestroy(); err != nil {
		return fmt.Errorf("could not terraform destroy - %w", err)
	}
	if err := d.cleanupResources(); err != nil {
		return fmt.Errorf("could not cleanup resources - %w", err)
	}
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

func (d *Destroy) cleanupResources() error {
	awsClient, err := aws.New()
	if err != nil {
		return err
	}
	tags := []aws.TagFilter{}
	for k, v := range d.ResourceTags {
		tags = append(tags, aws.TagFilter{Key: k, Values: []string{v}})
	}

	if err := awsClient.DeleteDynamodbTablesByTags(tags); err != nil {
		return err
	}
	return nil

}
