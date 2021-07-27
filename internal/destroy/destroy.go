package destroy

import (
	"fmt"

	"github.com/atoz-technology/mantil-backend/internal/assets"
	"github.com/atoz-technology/mantil-backend/internal/aws"
	"github.com/atoz-technology/mantil-backend/internal/mantil"
	"github.com/atoz-technology/mantil-backend/internal/terraform"
)

func Destroy(project *mantil.Project, path string) error {
	assets.StartServer()
	tf := terraform.New(path)
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
