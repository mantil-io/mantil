package destroy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/terraform"
	"github.com/mantil-io/mantil/workspace"
)

type DestroyRequest struct {
	ProjectName string
	StageName   string
}

type DestroyResponse struct{}

type Destroy struct {
	req        *DestroyRequest
	stage      *workspace.Stage
	bucketName string
	region     string
}

func New() *Destroy {
	return &Destroy{}
}

func (d *Destroy) Invoke(ctx context.Context, req *DestroyRequest) (*DestroyResponse, error) {
	if err := d.init(req); err != nil {
		return nil, err
	}
	return d.destroy()
}

func (d *Destroy) init(req *DestroyRequest) error {
	stage, err := workspace.LoadDeploymentState(req.ProjectName, req.StageName)
	if err != nil {
		return err
	}
	awsClient, err := aws.New()
	if err != nil {
		return err
	}
	bucketName, err := workspace.Bucket(awsClient)
	if err != nil {
		return err
	}
	d.req = req
	d.stage = stage
	d.bucketName = bucketName
	d.region = awsClient.Region()
	return nil
}

func (d *Destroy) destroy() (*DestroyResponse, error) {
	if err := d.terraformDestroy(); err != nil {
		return nil, fmt.Errorf("could not terraform destroy - %w", err)
	}
	if err := d.cleanupResources(); err != nil {
		return nil, fmt.Errorf("could not cleanup resources - %w", err)
	}
	if err := workspace.DeleteDeploymentState(d.req.ProjectName, d.req.StageName); err != nil {
		return nil, fmt.Errorf("could not delete stage %s - %w", d.req.StageName, err)
	}
	return &DestroyResponse{}, nil
}

func (d *Destroy) terraformDestroy() error {
	tf, err := terraform.Project(d.terraformProjectTemplateData())
	if err != nil {
		return err
	}
	return tf.Destroy()
}

func (d *Destroy) terraformProjectTemplateData() terraform.ProjectTemplateData {
	return terraform.ProjectTemplateData{
		Name:         d.req.ProjectName,
		Stage:        d.req.StageName,
		Bucket:       d.bucketName,
		BucketPrefix: workspace.DeploymentBucketPrefix(d.req.ProjectName, d.req.StageName),
		Region:       d.region,
	}
}

func (d *Destroy) cleanupResources() error {
	return workspace.CleanupResourcesFromDeployment(d.req.ProjectName, d.req.StageName)
}
