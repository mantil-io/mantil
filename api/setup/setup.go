package setup

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/assets"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/terraform"
)

type Setup struct{}

func New() *Setup {
	return &Setup{}
}

func (f *Setup) Invoke(ctx context.Context, req *dto.SetupRequest) (*dto.SetupResponse, error) {
	tf, err := terraform.New("mantil-setup")
	if err != nil {
		return nil, err
	}
	defer tf.Cleanup()
	rc := &config.RuntimeConfig{
		// TODO: sto ce mi ovaj version kada se nigdje ne koristi
		Version:         req.Version,
		FunctionsBucket: req.FunctionsBucket,
		FunctionsPath:   req.FunctionsPath,
	}
	return setup(tf, rc, req.PublicKey, req.Destroy)
}

func setup(tf *terraform.Terraform, rc *config.RuntimeConfig, publicKey string, destroy bool) (*dto.SetupResponse, error) {
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

	return &dto.SetupResponse{
		APIGatewayRestURL: restURL,
		APIGatewayWsURL:   wsURL,
	}, nil
}
