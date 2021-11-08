package setup

import (
	"context"
	"errors"
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
	if err := s.awsClient.S3().CreateBucket(req.BucketConfig.Name, req.ResourceTags); err != nil {
		return nil, err
	}
	if err := s.awsClient.S3().PutLifecycleRuleForPrefixExpire(req.BucketConfig.Name, req.BucketConfig.ExpirePrefix, req.BucketConfig.ExpireDays); err != nil {
		return nil, err
	}
	if err := s.setCloudwatchRoleForAPI("APIGatewayPushToCloudWatchLogsRole"); err != nil {
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

func (s *Setup) setCloudwatchRoleForAPI(name string) error {
	// check if role is already set
	role, err := s.awsClient.APIGAtewayCloudwatchRole()
	if err != nil {
		return err
	}
	// if role is set use its name so we can check if it exists or create it if it doesn't
	// otherwise leave default name
	if role != "" {
		name = role
	}
	arn, err := s.apiCloudwatchRoleArn(name)
	if err != nil {
		return err
	}
	// set role only if it wasn't already set
	if role == "" {
		return s.awsClient.SetAPIGatewayCloudwatchRole(arn)
	}
	return nil
}

func (s *Setup) apiCloudwatchRoleArn(name string) (string, error) {
	arn, err := s.awsClient.Role(name)
	if err == nil {
		return arn, nil
	}
	var rne *aws.RoleNotExistsError
	if errors.As(err, &rne) {
		arn, err = s.awsClient.CreateRole(name, apiCloudwatchRoleAssumePolicy, apiCloudwatchRolePolicy)
	}
	return arn, err
}

func (s *Setup) terraformCreate(req *dto.SetupRequest) (*dto.SetupResponse, error) {
	data := terraform.SetupTemplateData{
		Bucket:          req.BucketConfig.Name,
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
	cliRole := tf.Outputs["cli_role"]
	if url == "" {
		return nil, fmt.Errorf("can't find terraform output in %#v", tf.Outputs)
	}
	return &dto.SetupResponse{
		APIGatewayRestURL: url,
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

const apiCloudwatchRoleAssumePolicy = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const apiCloudwatchRolePolicy = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:DescribeLogGroups",
                "logs:DescribeLogStreams",
                "logs:PutLogEvents",
                "logs:GetLogEvents",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}`
