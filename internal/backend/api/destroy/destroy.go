package destroy

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/backend/assets"
	"github.com/mantil-io/mantil/internal/backend/terraform"
	"github.com/mantil-io/mantil/internal/mantil"
)

func Destroy(project *mantil.Project, tf *terraform.Terraform) error {
	assets.StartServer()
	aws, err := aws.New()
	if err != nil {
		return fmt.Errorf("could not initialize aws - %v", err)
	}
	if err := tf.ApplyForProject(project, aws, true); err != nil {
		return fmt.Errorf("could not terraform destroy - %v", err)
	}
	if err := mantil.DeleteProject(project, aws); err != nil {
		return fmt.Errorf("could not delete project %s - %v", project.Name, err)
	}
	return nil
}
