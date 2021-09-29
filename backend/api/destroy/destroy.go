package destroy

import (
	"fmt"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/terraform"
)

func Destroy(projectName string, stage *config.Stage, awsClient *aws.AWS, rc *config.RuntimeConfig) error {
	_, err := terraformDestroy(projectName, stage, awsClient)
	if err != nil {
		return fmt.Errorf("could not terraform destroy - %v", err)
	}
	if err := config.DeleteDeploymentState(projectName, stage.Name); err != nil {
		return fmt.Errorf("could not delete project %s - %v", projectName, err)
	}
	return nil
}

func terraformDestroy(projectName string, stage *config.Stage, awsClient *aws.AWS) (*terraform.Terraform, error) {
	bucket, err := config.Bucket(awsClient)
	if err != nil {
		return nil, err
	}
	data := terraform.ProjectTemplateData{
		Name:         projectName,
		Bucket:       bucket,
		BucketPrefix: config.DeploymentBucketPrefix(projectName, stage.Name),
		Region:       awsClient.Region(),
	}
	tf, err := terraform.Project(data)
	if err != nil {
		return nil, err
	}
	return tf, tf.Destroy()
}
