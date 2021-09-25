package setup

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/backend/assets"
	"github.com/mantil-io/mantil/internal/backend/terraform"
	"github.com/mantil-io/mantil/internal/config"
)

type SetupOutput struct {
	RestURL string
	WsURL   string
}

func Setup(tf *terraform.Terraform, rc *config.RuntimeConfig, publicKey string, destroy bool) (*SetupOutput, error) {
	assets.StartServer()
	awsClient, err := aws.New()
	if err != nil {
		return nil, fmt.Errorf("error initializing AWS client - %v", err)
	}
	bucketName, err := config.Bucket(awsClient)
	if err != nil {
		return nil, fmt.Errorf("error creating bucket - %v", bucketName)
	}
	bucketExists, err := awsClient.S3BucketExists(bucketName)
	if err != nil {
		return nil, fmt.Errorf("error checking if bucket exists - %v", err)
	}
	if !bucketExists {
		if err := awsClient.CreateS3Bucket(bucketName, awsClient.Region()); err != nil {
			return nil, fmt.Errorf("error creating terraform bucket - %v", err)
		}

	}
	if err := tf.RenderSetupTemplate(bucketName, rc, publicKey, awsClient); err != nil {
		return nil, err
	}
	if err := tf.Apply(destroy); err != nil {
		return nil, err
	}
	if destroy {
		if err := awsClient.EmptyS3Bucket(bucketName); err != nil {
			return nil, fmt.Errorf("error emptying terraform bucket - %v", err)
		}
		if err := awsClient.DeleteS3Bucket(bucketName); err != nil {
			return nil, fmt.Errorf("error deleting terraform bucket - %v", err)
		}
	} else {
		if err := config.SaveRuntimeConfig(awsClient, rc); err != nil {
			return nil, fmt.Errorf("error saving mantil version")
		}
	}
	restURL, err := tf.Output("url", true)
	if err != nil {
		return nil, fmt.Errorf("error reading api gateway rest url - %v", err)
	}
	wsURL, err := tf.Output("ws_url", true)
	if err != nil {
		return nil, fmt.Errorf("error reading api gateway ws url - %v", err)
	}
	return &SetupOutput{
		RestURL: restURL,
		WsURL:   wsURL,
	}, nil
}
