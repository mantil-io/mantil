package setup

import (
	"fmt"
	"os"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/workspace"
)

type Args struct {
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

func (a *Args) ParseArgs(args []string) {
	if len(args) == 0 {
		a.AccountName = DefaultAccountName()
		return
	}
	a.AccountName = args[0]
}

func (a *Args) validate() error {
	if a.AccessKeyID != "" || a.SecretAccessKey != "" {
		return a.validateAccessKeys()
	}
	if a.UseEnv {
		return a.readFromEnv()
	}
	if a.Profile != "" {
		return nil
	}
	return fmt.Errorf("AWS credentials not provided")
}

func (a *Args) validateAccessKeys() error {
	if a.AccessKeyID == "" {
		return fmt.Errorf("aws-access-key-id not provided, must be used with the aws-secret-access-key and aws-region")
	}
	if a.SecretAccessKey == "" {
		return fmt.Errorf("aws-secret-access-key not provided, must be used with the aws-access-key-id and aws-region")
	}
	if a.Region == "" {
		return fmt.Errorf("aws-region not provided, must be used with aws-access-key-id and aws-secret-access-key")
	}
	return nil
}

const (
	accessKeyIDEnv     = "AWS_ACCESS_KEY_ID"
	secretAccessKeyEnv = "AWS_SECRET_ACCESS_KEY"
	regionEnv          = "AWS_DEFAULT_REGION"
)

func (a *Args) readFromEnv() error {
	errf := func(env string) error {
		return fmt.Errorf("%s environment variable not provided", env)
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

func (a *Args) awsConnect() (*aws.AWS, error) {
	cli, err := a.awsClient()
	if err != nil {
		return nil, err
	}
	a.AccountID = cli.AccountID()
	a.Region = cli.Region()
	return cli, nil
}

func (a *Args) awsClient() (*aws.AWS, error) {
	if a.Profile != "" {
		return aws.NewFromProfile(a.Profile)
	}
	return aws.NewWithCredentials(a.AccessKeyID, a.SecretAccessKey, "", a.Region)
}
