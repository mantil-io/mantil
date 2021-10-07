package setup

import (
	"fmt"
	"os"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/workspace"
)

type Flags struct {
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

func (f *Flags) ParseArgs(args []string) {
	if len(args) == 0 {
		f.AccountName = DefaultAccountName()
		return
	}
	f.AccountName = args[0]
}

func (f *Flags) validate() error {
	if f.AccessKeyID != "" || f.SecretAccessKey != "" {
		return f.validateAccessKeys()
	}
	if f.UseEnv {
		return f.readFromEnv()
	}
	if f.Profile != "" {
		return nil
	}
	return fmt.Errorf("AWS credentials not provided")
}

func (f *Flags) validateAccessKeys() error {
	if f.AccessKeyID == "" {
		return fmt.Errorf("aws-access-key-id not provided, must be used with the aws-secret-access-key and aws-region")
	}
	if f.SecretAccessKey == "" {
		return fmt.Errorf("aws-secret-access-key not provided, must be used with the aws-access-key-id and aws-region")
	}
	if f.Region == "" {
		return fmt.Errorf("aws-region not provided, must be used with aws-access-key-id and aws-secret-access-key")
	}
	return nil
}

const (
	accessKeyIDEnv     = "AWS_ACCESS_KEY_ID"
	secretAccessKeyEnv = "AWS_SECRET_ACCESS_KEY"
	regionEnv          = "AWS_DEFAULT_REGION"
)

func (f *Flags) readFromEnv() error {
	errf := func(env string) error {
		return fmt.Errorf("%s environment variable not provided", env)
	}
	f.AccessKeyID, _ = os.LookupEnv(accessKeyIDEnv)
	if f.AccessKeyID == "" {
		return errf(accessKeyIDEnv)
	}
	f.SecretAccessKey, _ = os.LookupEnv(secretAccessKeyEnv)
	if f.SecretAccessKey == "" {
		return errf(secretAccessKeyEnv)
	}
	f.Region, _ = os.LookupEnv(regionEnv)
	if f.Region == "" {
		return errf(regionEnv)
	}
	return nil
}

func (f *Flags) awsConnect() (*aws.AWS, error) {
	cli, err := f.awsClient()
	if err != nil {
		return nil, err
	}
	f.AccountID = cli.AccountID()
	f.Region = cli.Region()
	return cli, nil
}

func (f *Flags) awsClient() (*aws.AWS, error) {
	if f.Profile != "" {
		return aws.NewFromProfile(f.Profile)
	}
	return aws.NewWithCredentials(f.AccessKeyID, f.SecretAccessKey, "", f.Region)
}
