package destroy

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/backend/assets"
	"github.com/mantil-io/mantil/internal/backend/terraform"
	"github.com/mantil-io/mantil/internal/mantil"
)

func Destroy(project *mantil.Project, stage string) error {
	assets.StartServer()
	aws, err := aws.New()
	if err != nil {
		return fmt.Errorf("could not initialize aws - %v", err)
	}
	if stage != "" {
		if err := destroyStage(project, stage, aws); err != nil {
			return fmt.Errorf("could not terraform destroy - %v", err)
		}
		if err := mantil.DeleteProjectStage(project, stage, aws); err != nil {
			return fmt.Errorf("could not delete project %s - %v", project.Name, err)
		}
		project.RemoveStage(stage)
		mantil.SaveProjectS3(project)
	} else {
		for _, s := range project.Stages {
			if err := destroyStage(project, s.Name, aws); err != nil {
				return fmt.Errorf("could not terraform destroy - %v", err)
			}
		}
		if err := mantil.DeleteProject(project, aws); err != nil {
			return fmt.Errorf("could not delete project %s - %v", project.Name, err)
		}
	}
	return nil
}

func destroyStage(project *mantil.Project, stage string, aws *aws.AWS) error {
	tf, err := terraform.New(fmt.Sprintf("%s-%s", project.Name, stage))
	if err != nil {
		return err
	}
	defer tf.Cleanup()
	if err := tf.ApplyForProject(project, stage, aws, true); err != nil {
		return fmt.Errorf("could not terraform destroy - %v", err)
	}
	return nil
}
