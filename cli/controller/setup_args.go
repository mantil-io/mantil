package controller

import (
	"os"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/domain"
)

const (
	accessKeyIDEnv     = "AWS_ACCESS_KEY_ID"
	secretAccessKeyEnv = "AWS_SECRET_ACCESS_KEY"
	regionEnv          = "AWS_DEFAULT_REGION"
	sessionTokenEnv    = "AWS_SESSION_TOKEN"
)

type SetupArgs struct {
	AccessKeyID         string
	SecretAccessKey     string
	SessionToken        string
	Region              string
	Profile             string
	UseEnv              bool
	NodeName            string
	DryRun              bool
	AccountID           string
	credentialsProvider int
	Force               bool
	Yes                 bool
}

func DefaultNodeName() string { return domain.DefaultNodeName }

func (a *SetupArgs) ParseArgs(args []string) {
	if len(args) == 0 {
		a.NodeName = DefaultNodeName()
		return
	}
	a.NodeName = args[0]
}

func (a *SetupArgs) validate() error {
	if a.AccessKeyID != "" || a.SecretAccessKey != "" {
		a.credentialsProvider = domain.AWSCredentialsByArguments
		return a.validateAccessKeys()
	}
	if a.UseEnv {
		a.credentialsProvider = domain.AWSCredentialsByEnv
		return a.readFromEnv()
	}
	if a.Profile != "" {
		a.credentialsProvider = domain.AWSCredentialsByProfile
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
	a.SessionToken, _ = os.LookupEnv(sessionTokenEnv)
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
	return aws.NewWithCredentials(a.AccessKeyID, a.SecretAccessKey, a.SessionToken, a.Region)
}
