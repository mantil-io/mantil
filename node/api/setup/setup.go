package setup

import (
	"context"
	"errors"
	"fmt"

	"github.com/mantil-io/mantil/kit/aws"
	"github.com/mantil-io/mantil/kit/token"
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
	if err := s.setCloudwatchRoleForAPI(req.APIGatewayLogsRole); err != nil {
		return nil, err
	}
	out, err := s.terraformCreate(req)
	if err != nil {
		return nil, err
	}
	return out, err
}

func (s *Setup) Upgrade(ctx context.Context, req *dto.SetupRequest) error {
	if err := s.init(); err != nil {
		return err
	}
	_, err := s.terraformCreate(req)
	if err != nil {
		return err
	}
	return err
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
	role, err := s.awsClient.APIGAtewayCloudwatchRole()
	if err != nil {
		return err
	}
	if role != "" {
		exists, err := s.awsClient.RoleExists(role)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
	}
	arn, err := s.apiCloudwatchRoleArn(name)
	if err != nil {
		return err
	}
	return s.awsClient.SetAPIGatewayCloudwatchRole(arn)
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
	publicKey, privateKey, err := token.KeyPair()
	if err != nil {
		return nil, err
	}
	data := terraform.SetupTemplateData{
		Bucket:          req.BucketConfig.Name,
		Region:          s.awsClient.Region(),
		FunctionsBucket: req.FunctionsBucket,
		FunctionsPath:   req.FunctionsPath,
		ResourceSuffix:  req.ResourceSuffix,
		NamingTemplate:  req.NamingTemplate,
		AuthEnv:         req.AuthEnv,
		ResourceTags:    req.ResourceTags,
		PublicKey:       publicKey,
		PrivateKey:      privateKey,
	}
	tf, err := terraform.Setup(data)
	if err != nil {
		return nil, err
	}
	if err := tf.Create(); err != nil {
		return nil, err
	}
	url, err := tf.Output("url")
	if err != nil {
		return nil, err
	}
	cliRole, err := tf.Output("cli_role")
	if err != nil {
		return nil, err
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
