package destroy

import (
	"fmt"

	"github.com/mantil-io/mantil/assets"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/terraform"
)

func Destroy(projectName string, stage *config.Stage, tf *terraform.Terraform, awsClient *aws.AWS, rc *config.RuntimeConfig) error {
	assets.StartServer()
	if err := tf.ApplyForProject(projectName, stage, awsClient, rc, true); err != nil {
		return fmt.Errorf("could not terraform destroy - %v", err)
	}
	if err := config.DeleteDeploymentState(projectName, stage.Name); err != nil {
		return fmt.Errorf("could not delete project %s - %v", projectName, err)
	}
	return nil
}
