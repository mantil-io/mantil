package destroy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/terraform"
	"github.com/mantil-io/mantil/workspace"
)

type Destroy struct {
	dto.DestroyRequest
}

func New() *Destroy {
	return &Destroy{}
}

func (d *Destroy) Invoke(ctx context.Context, req *dto.DestroyRequest) error {
	d.DestroyRequest = *req

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
		BucketPrefix: workspace.StageBucketPrefix(d.ProjectName, d.StageName),
	}
}

func (d *Destroy) cleanupResources() error {
	awsClient, err := aws.New()
	if err != nil {
		return err
	}
	tags := []aws.TagFilter{
		{Key: workspace.EnvProjectName, Values: []string{d.ProjectName}},
		{Key: workspace.EnvStageName, Values: []string{d.StageName}},
	}
	if err := awsClient.DeleteDynamodbTablesByTags(tags); err != nil {
		return err
	}
	return nil

}
