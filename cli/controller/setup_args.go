package controller

import (
	"os"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/workspace"
)

const (
	accessKeyIDEnv     = "AWS_ACCESS_KEY_ID"
	secretAccessKeyEnv = "AWS_SECRET_ACCESS_KEY"
	regionEnv          = "AWS_DEFAULT_REGION"
)

type SetupArgs struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Profile         string
	UseEnv          bool
	AccountName     string
	DryRun          bool
	Override        bool
	AccountID       string
}

func DefaultAccountName() string { return workspace.DefaultAccountName }

func (a *SetupArgs) ParseArgs(args []string) {
	if len(args) == 0 {
		a.AccountName = DefaultAccountName()
		return
	}
	a.AccountName = args[0]
}

func (a *SetupArgs) validate() error {
	if a.AccessKeyID != "" || a.SecretAccessKey != "" {
		return a.validateAccessKeys()
	}
	if a.UseEnv {
		return a.readFromEnv()
	}
	if a.Profile != "" {
		return nil
	}
	return log.Wrap(NewArgumentError("AWS credentials not provided"))
}

func (a *SetupArgs) validateAccessKeys() error {
	if a.AccessKeyID == "" {
		return log.Wrap(NewArgumentError("aws-access-key-id not provided, must be used with the aws-secret-access-key and aws-region"))
	}
	if a.SecretAccessKey == "" {
		return log.Wrap(NewArgumentError("aws-secret-access-key not provided, must be used with the aws-access-key-id and aws-region"))
	}
	if a.Region == "" {
		return log.Wrap(NewArgumentError("aws-region not provided, must be used with aws-access-key-id and aws-secret-access-key"))
	}
	return nil
}

func (a *SetupArgs) readFromEnv() error {
	errf := func(env string) error {
		return NewArgumentError("%s environment variable not provided", env)
	}
	a.AccessKeyID, _ = os.LookupEnv(accessKeyIDEnv)
	if a.AccessKeyID == "" {
		return errf(accessKeyIDEnv)
	}
	a.SecretAccessKey, _ = os.LookupEnv(secretAccessKeyEnv)
	if a.SecretAccessKey == "" {
		return errf(secretAccessKeyEnv)
	}
	a.Region, _ = os.LookupEnv(regionEnv)
	if a.Region == "" {
		return errf(regionEnv)
	}
	return nil
}

func (a *SetupArgs) awsConnect() (*aws.AWS, error) {
	cli, err := a.awsClient()
	if err != nil {
		return nil, err
	}
	a.AccountID = cli.AccountID()
	a.Region = cli.Region()
	return cli, nil
}

func (a *SetupArgs) awsClient() (*aws.AWS, error) {
	if a.Profile != "" {
		return aws.NewFromProfile(a.Profile)
	}
	return aws.NewWithCredentials(a.AccessKeyID, a.SecretAccessKey, "", a.Region)
}
