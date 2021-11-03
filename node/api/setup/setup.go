package setup

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/node/dto"
	"github.com/mantil-io/mantil/node/terraform"
)

type Setup struct {
	awsClient *aws.AWS
}

func New() *Setup {
	return &Setup{}
}

func (s *Setup) Destroy(ctx context.Context, req *dto.SetupDestroyRequest) error {
	if err := s.init(); err != nil {
		return err
	}
	if err := s.terraformDestroy(req); err != nil {
		return err
	}
	if err := s.awsClient.S3().DeleteBucket(req.Bucket); err != nil {
		return err
	}
	return nil
}

func (s *Setup) Create(ctx context.Context, req *dto.SetupRequest) (*dto.SetupResponse, error) {
	if err := s.init(); err != nil {
		return nil, err
	}
	if err := s.awsClient.S3().CreateBucket(req.Bucket, req.ResourceTags); err != nil {
		return nil, err
	}
	out, err := s.terraformCreate(req)
	if err != nil {
		return nil, err
	}
	return out, err
}

func (s *Setup) init() error {
	awsClient, err := aws.New()
	if err != nil {
		return fmt.Errorf("error initializing AWS client - %w", err)
	}
	s.awsClient = awsClient
	return nil
}

func (s *Setup) terraformCreate(req *dto.SetupRequest) (*dto.SetupResponse, error) {
	data := terraform.SetupTemplateData{
		Bucket:          req.Bucket,
		Region:          s.awsClient.Region(),
		FunctionsBucket: req.FunctionsBucket,
		FunctionsPath:   req.FunctionsPath,
		ResourceSuffix:  req.ResourceSuffix,
		AuthEnv:         req.AuthEnv,
		ResourceTags:    req.ResourceTags,
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

func (s *Setup) terraformDestroy(req *dto.SetupDestroyRequest) error {
	data := terraform.SetupTemplateData{
		Bucket: req.Bucket,
		Region: s.awsClient.Region(),
	}
	tf, err := terraform.Setup(data)
	if err != nil {
		return err
	}
	return tf.Destroy()
}
