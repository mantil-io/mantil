package bootstrap

import (
	"fmt"

	"github.com/atoz-technology/mantil-backend/internal/assets"
	"github.com/atoz-technology/mantil-backend/internal/aws"
	"github.com/atoz-technology/mantil-backend/internal/terraform"
)

func Bootstrap(path string, destroy bool) error {
	assets.StartServer()
	awsClient, err := aws.New()
	if err != nil {
		return fmt.Errorf("error initializing AWS client - %v", err)
	}
	accoundID, err := awsClient.AccountID()
	if err != nil {
		return fmt.Errorf("error fetching AWS account ID - %v", err)
	}
	bucketName := fmt.Sprintf("mantil-bootstrap-%v", accoundID)
	if !destroy {
		if err := awsClient.CreateS3Bucket(bucketName, awsClient.Region()); err != nil {
			return fmt.Errorf("error creating terraform bucket - %v", err)
		}
	}
	tf := terraform.New(path)
	if err := tf.ApplyForBootstrap(bucketName, destroy); err != nil {
		return err
	}
	if destroy {
		if err := awsClient.EmptyS3Bucket(bucketName); err != nil {
			return fmt.Errorf("error emptying terraform bucket - %v", err)
		}
		if err := awsClient.DeleteS3Bucket(bucketName); err != nil {
			return fmt.Errorf("error deleting terraform bucket - %v", err)
		}
	}
	return nil
}
