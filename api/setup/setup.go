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

type Setup struct {
	tf         *terraform.Terraform
	req        *dto.SetupRequest
	awsClient  *aws.AWS
	bucketName string
}

func New() *Setup {
	return &Setup{}
}

func (s *Setup) Invoke(ctx context.Context, req *dto.SetupRequest) (*dto.SetupResponse, error) {
	if err := s.init(req); err != nil {
		return nil, err
	}
	defer s.cleanup()

	if req.Destroy {
		return nil, s.destroy()
	}

	if err := s.create(); err != nil {
		return nil, err
	}
	return s.readTerraformOutput()
}

func (s *Setup) destroy() error {
	if err := s.terraformApply(); err != nil {
		return err
	}
	if err := s.deleteBucket(); err != nil {
		return err
	}
	return nil
}

func (s *Setup) create() error {
	if err := s.createBucket(); err != nil {
		return err
	}
	if err := s.terraformApply(); err != nil {
		return err
	}
	if err := s.saveConfig(); err != nil {
		return err
	}
	return nil
}

func (s *Setup) init(req *dto.SetupRequest) error {
	tf, err := terraform.New("mantil-setup")
	if err != nil {
		return err
	}
	awsClient, err := aws.New()
	if err != nil {
		return fmt.Errorf("error initializing AWS client - %w", err)
	}
	bucketName, err := config.Bucket(awsClient)
	if err != nil {
		return err
	}
	s.tf = tf
	s.awsClient = awsClient
	s.req = req
	s.bucketName = bucketName
	return nil
}

func (s *Setup) saveConfig() error {
	return config.SaveRuntimeConfig(s.awsClient, &config.RuntimeConfig{
		// TODO: sto ce mi ovaj version kada se nigdje ne koristi
		Version:         s.req.Version,
		FunctionsBucket: s.req.FunctionsBucket,
		FunctionsPath:   s.req.FunctionsPath,
	})
}

func (s *Setup) cleanup() {
	if s.tf != nil {
		s.tf.Cleanup()
	}
}

func (s *Setup) terraformApply() error {
	assets.StartServer()
	data := terraform.SetupTemplateData{
		Bucket:          s.bucketName,
		BucketPrefix:    config.SetupBucketPrefix,
		FunctionsBucket: s.req.FunctionsBucket,
		FunctionsPath:   s.req.FunctionsPath,
		Region:          s.awsClient.Region(),
		PublicKey:       s.req.PublicKey,
	}
	if err := s.tf.RenderSetupTemplate(data); err != nil {
		return err
	}
	if err := s.tf.Apply(s.req.Destroy); err != nil {
		return err
	}
	return nil
}

func (s *Setup) readTerraformOutput() (*dto.SetupResponse, error) {
	restURL, err := s.tf.Output("url", true)
	if err != nil {
		return nil, fmt.Errorf("error reading api gateway rest url - %v", err)
	}
	wsURL, err := s.tf.Output("ws_url", true)
	if err != nil {
		return nil, fmt.Errorf("error reading api gateway ws url - %v", err)
	}
	return &dto.SetupResponse{
		APIGatewayRestURL: restURL,
		APIGatewayWsURL:   wsURL,
	}, nil
}

func (s *Setup) deleteBucket() error {
	if err := s.awsClient.EmptyS3Bucket(s.bucketName); err != nil {
		return fmt.Errorf("error emptying bucket %s - %w", s.bucketName, err)
	}
	if err := s.awsClient.DeleteS3Bucket(s.bucketName); err != nil {
		return fmt.Errorf("error deleting bucket %s - %w", s.bucketName, err)
	}
	return nil
}

func (s *Setup) createBucket() error {
	exists, err := s.awsClient.S3BucketExists(s.bucketName)
	if err != nil {
		return fmt.Errorf("error checking if bucket %s exists - %w", s.bucketName, err)
	}
	if exists {
		return nil
	}
	if err := s.awsClient.CreateS3Bucket(s.bucketName, s.awsClient.Region()); err != nil {
		return fmt.Errorf("error creating bucket %s - %w", s.bucketName, err)
	}
	return nil
}
