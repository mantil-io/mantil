package setup

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/terraform"
)

type Setup struct {
	req       *dto.SetupRequest
	awsClient *aws.AWS
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
	s.awsClient = awsClient
	s.req = req
	return nil
}

func (s *Setup) terraformCreate() (*dto.SetupResponse, error) {
	data := terraform.SetupTemplateData{
		Bucket:          s.req.Bucket,
		Region:          s.awsClient.Region(),
		FunctionsBucket: s.req.FunctionsBucket,
		FunctionsPath:   s.req.FunctionsPath,
		PublicKey:       s.req.PublicKey,
		ResourceSuffix:  s.req.ResourceSuffix,
		ResourceTags:    s.req.ResourceTags,
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
	cliRole := tf.Outputs["cli_role"]
	if url == "" || wsURL == "" {
		return nil, fmt.Errorf("can't find terraform output in %#v", tf.Outputs)
	}
	return &dto.SetupResponse{
		APIGatewayRestURL: url,
		APIGatewayWsURL:   wsURL,
		CliRole:           cliRole,
	}, nil
}

func (s *Setup) terraformDestroy() error {
	data := terraform.SetupTemplateData{
		Bucket: s.req.Bucket,
		Region: s.awsClient.Region(),
	}
	tf, err := terraform.Setup(data)
	if err != nil {
		return err
	}
	return tf.Destroy()
}

func (s *Setup) deleteBucket() error {
	if err := s.awsClient.EmptyS3Bucket(s.req.Bucket); err != nil {
		return fmt.Errorf("error emptying bucket %s - %w", s.req.Bucket, err)
	}
	if err := s.awsClient.DeleteS3Bucket(s.req.Bucket); err != nil {
		return fmt.Errorf("error deleting bucket %s - %w", s.req.Bucket, err)
	}
	return nil
}

func (s *Setup) createBucket() error {
	exists, err := s.awsClient.S3BucketExists(s.req.Bucket)
	if err != nil {
		return fmt.Errorf("error checking if bucket %s exists - %w", s.req.Bucket, err)
	}
	if exists {
		return nil
	}
	if err := s.awsClient.CreateS3Bucket(s.req.Bucket, s.awsClient.Region()); err != nil {
		return fmt.Errorf("error creating bucket %s - %w", s.req.Bucket, err)
	}
	if err := s.awsClient.TagS3Bucket(s.req.Bucket, s.req.ResourceTags); err != nil {
		return fmt.Errorf("error tagging bucket %s - %w", s.req.Bucket, err)
	}
	return nil
}
