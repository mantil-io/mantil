package destroy

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/backend/assets"
	"github.com/mantil-io/mantil/internal/backend/terraform"
	"github.com/mantil-io/mantil/internal/config"
)

func Destroy(project *config.Project, stage string, tf *terraform.Terraform, awsClient *aws.AWS, rc *config.RuntimeConfig) error {
	assets.StartServer()
	if stage != "" {
		if err := destroyStage(project, stage, tf, awsClient, rc); err != nil {
			return fmt.Errorf("could not terraform destroy - %v", err)
		}
		if err := config.DeleteProjectStage(project, stage, awsClient); err != nil {
			return fmt.Errorf("could not delete project %s - %v", project.Name, err)
		}
		project.RemoveStage(stage)
		config.SaveProjectS3(project)
	} else {
		for _, s := range project.Stages {
			if err := destroyStage(project, s.Name, tf, awsClient, rc); err != nil {
				return fmt.Errorf("could not terraform destroy - %v", err)
			}
		}
		if err := config.DeleteProject(project, awsClient); err != nil {
			return fmt.Errorf("could not delete project %s - %v", project.Name, err)
		}
	}
	return nil
}

func destroyStage(project *config.Project, stage string, tf *terraform.Terraform, aws *aws.AWS, rc *config.RuntimeConfig) error {
	if err := tf.ApplyForProject(project, stage, aws, rc, true); err != nil {
		return fmt.Errorf("could not terraform destroy - %v", err)
	}
	return nil
}
