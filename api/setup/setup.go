package setup

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/terraform"
	"github.com/mantil-io/mantil/workspace"
)

type Setup struct {
	req        *dto.SetupRequest
	awsClient  *aws.AWS
	bucketName string
}

func New() *Setup {
	return &Setup{}
}

func (s *Setup) Invoke(ctx context.Context, req *dto.SetupRequest) (*dto.SetupResponse, error) {
	if err := s.init(req, nil); err != nil {
		return nil, err
	}
	if req.Destroy {
		return nil, s.destroy()
	}
	return s.create()
}

func (s *Setup) destroy() error {
	if err := s.terraformDestroy(); err != nil {
		return err
	}
	if err := s.deleteBucket(); err != nil {
		return err
	}
	return nil
}

func (s *Setup) create() (*dto.SetupResponse, error) {
	if err := s.createBucket(); err != nil {
		return nil, err
	}
	out, err := s.terraformCreate()
	if err != nil {
		return nil, err
	}
	if err := s.saveConfig(); err != nil {
		return nil, err
	}
	return out, err
}

func (s *Setup) init(req *dto.SetupRequest, awsClient *aws.AWS) error {
	if awsClient == nil {
		var err error
		awsClient, err = aws.New()
		if err != nil {
			return fmt.Errorf("error initializing AWS client - %w", err)
		}
	}
	bucketName, err := workspace.Bucket(awsClient)
	if err != nil {
		return err
	}
	s.awsClient = awsClient
	s.req = req
	s.bucketName = bucketName
	return nil
}

func (s *Setup) saveConfig() error {
	return workspace.SaveRuntimeConfig(s.awsClient, &workspace.RuntimeConfig{
		Version:         s.req.Version,
		FunctionsBucket: s.req.FunctionsBucket,
		FunctionsPath:   s.req.FunctionsPath,
	})
}

func (s *Setup) terraformCreate() (*dto.SetupResponse, error) {
	data := terraform.SetupTemplateData{
		Bucket:          s.bucketName,
		Region:          s.awsClient.Region(),
		FunctionsBucket: s.req.FunctionsBucket,
		FunctionsPath:   s.req.FunctionsPath,
		PublicKey:       s.req.PublicKey,
	}
	tf, err := terraform.Setup(data)
	if err != nil {
		return nil, err
	}
	if err := tf.Create(); err != nil {
		return nil, err
	}
	url := tf.Outputs["url"]
	wsURL := tf.Outputs["ws_url"]
	if url == "" || wsURL == "" {
		return nil, fmt.Errorf("can't find terraform output in %#v", tf.Outputs)
	}
	return &dto.SetupResponse{
		APIGatewayRestURL: url,
		APIGatewayWsURL:   wsURL,
	}, nil
}

func (s *Setup) terraformDestroy() error {
	data := terraform.SetupTemplateData{
		Bucket: s.bucketName,
		Region: s.awsClient.Region(),
	}
	tf, err := terraform.Setup(data)
	if err != nil {
		return err
	}
	return tf.Destroy()
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
