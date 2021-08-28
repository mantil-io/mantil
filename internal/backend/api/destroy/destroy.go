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
	if err := tf.ApplyForProject(project, true); err != nil {
		return fmt.Errorf("could not terraform destroy - %v", err)
	}
	aws, err := aws.New()
	if err != nil {
		return fmt.Errorf("could not initialize aws - %v", err)
	}
	bucketName := project.Bucket
	bucketExists, _ := aws.S3BucketExists(bucketName)
	if bucketExists {
		err = aws.DeleteS3Bucket(bucketName)
		if err != nil {
			return fmt.Errorf("could not delete bucket %s - %v", bucketName, err)
		}
	}
	return nil
}
