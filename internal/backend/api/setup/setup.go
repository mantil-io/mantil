package setup

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/backend/assets"
	"github.com/mantil-io/mantil/internal/backend/terraform"
)

func Setup(tf *terraform.Terraform, destroy bool) (string, error) {
	assets.StartServer()
	awsClient, err := aws.New()
	if err != nil {
		return "", fmt.Errorf("error initializing AWS client - %v", err)
	}
	accoundID, err := awsClient.AccountID()
	if err != nil {
		return "", fmt.Errorf("error fetching AWS account ID - %v", err)
	}
	bucketName := fmt.Sprintf("mantil-setup-%v", accoundID)
	if !destroy {
		if err := awsClient.CreateS3Bucket(bucketName, awsClient.Region()); err != nil {
			return "", fmt.Errorf("error creating terraform bucket - %v", err)
		}
	}
	if err := tf.ApplyForSetup(bucketName, destroy); err != nil {
		return "", err
	}
	if destroy {
		if err := awsClient.EmptyS3Bucket(bucketName); err != nil {
			return "", fmt.Errorf("error emptying terraform bucket - %v", err)
		}
		if err := awsClient.DeleteS3Bucket(bucketName); err != nil {
			return "", fmt.Errorf("error deleting terraform bucket - %v", err)
		}
	}
	url, err := tf.Output("url", true)
	if err != nil {
		return "", fmt.Errorf("error reading api gateway url - %v", err)
	}
	return url, nil
}
