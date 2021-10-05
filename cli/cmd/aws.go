package cmd

import (
	"fmt"
	"os"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/cmd/setup"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
	"github.com/spf13/cobra"
)

func newAwsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: "AWS account subcommand",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(newAwsInstallCommand())
	cmd.AddCommand(newAwsUninstallCommand())
	return cmd
}

func newAwsInstallCommand() *cobra.Command {
	var f awsFlags
	cmd := &cobra.Command{
		Use:   "install [account-name]",
		Short: "Install Mantil into AWS account",
		Long: fmt.Sprintf(`Install Mantil into AWS account

Command will install backend services into AWS account.
You must provide credentials for Mantil to access your AWS account.
%s
Argument account-name is for referencing that account in Mantil.
If not provided default name %s will be used for the first account.

There is --dry-run flag which will show you what credentials will be used
and what account will be managed by command.
`, credentialsHelp(), workspace.DefaultAccountName),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.ParseArgs(args)
			stp, err := newSetup(&f)
			if err != nil {
				return log.Wrap(err)
			}
			if f.DryRun {
				showAwsDryRunInfo(&f)
				return nil
			}
			return stp.Create()
		},
	}
	bindAwsInstallFlags(cmd, &f)
	cmd.Flags().BoolVar(&f.Override, "override", false, "force override access tokens on already installed account")
	return cmd
}

func newAwsUninstallCommand() *cobra.Command {
	var f awsFlags
	cmd := &cobra.Command{
		Use:   "uninstall [account-name]",
		Short: "Uninstall Mantil from AWS account",
		Long: fmt.Sprintf(`Uninstall Mantil from AWS account

Command will remove backend services from AWS account.
You must provide credentials for Mantil to access your AWS account.
%s
Argument account-name is Mantil account reference.

There is --dry-run flag which will show you what credentials will be used
and what account will be managed by command.
`, credentialsHelp()),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.ParseArgs(args)
			stp, err := newSetup(&f)
			if err != nil {
				return log.Wrap(err)
			}
			if f.DryRun {
				showAwsDryRunInfo(&f)
				return nil
			}
			return stp.Destroy()
		},
	}
	bindAwsInstallFlags(cmd, &f)
	return cmd
}

func newSetup(f *awsFlags) (*setup.Cmd, error) {
	if err := f.Validate(); err != nil {
		return nil, err
	}
	cli, err := f.awsConnect()
	if err != nil {
		return nil, log.WithUserMessage(err, "invalid AWS access credentials")
	}
	return setup.New(cli, f.AccountName, f.Override), nil
}

func credentialsHelp() string {
	return `There are few ways to provide credentials:

1. specifiy access keys as arguments:
   $ mantil aws install --aws-access-key-id=AKIAIOSFODNN7EXAMPLE --aws-secret-access-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --aws-region=us-east-1

2. read access keys from environment variables:
   $ export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
   $ export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
   $ export AWS_DEFAULT_REGION=us-east-1
   $ mantil aws install --aws-env

reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html

3. use your named AWS profile form ~/.aws/config
   $ mantil aws install --aws-profile=my-named-profile

reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html
`
}

func bindAwsInstallFlags(cmd *cobra.Command, f *awsFlags) {
	cmd.Flags().StringVar(&f.AccessKeyID, "aws-access-key-id", "", "access key ID for the AWS account, must be used with the aws-secret-access-key and aws-region flags")
	cmd.Flags().StringVar(&f.SecretAccessKey, "aws-secret-access-key", "", "secret access key for the AWS account, must be used with the aws-access-key-id and aws-region flags")
	cmd.Flags().StringVar(&f.Region, "aws-region", "", "region for the AWS account, must be used with and aws-access-key-id and aws-secret-access-key flags")
	cmd.Flags().BoolVar(&f.UseEnv, "aws-env", false, "use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION environment variables for AWS authentication")
	cmd.Flags().StringVar(&f.Profile, "aws-profile", "", "use the given profile for AWS authentication")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "don't start install/uninstall just show what credentials will be used")
}

func showAwsDryRunInfo(f *awsFlags) {
	if f.Profile != "" {
		ui.Info(`Command will use AWS profile %s defined in your AWS configuration file (~/.aws/config)`, f.Profile)
	} else {
		ui.Info(`Command will use AWS credentials:
  aws-access-key-id: %s
  aws-secret-access-key: %s
  aws-region: %s`, f.AccessKeyID, f.SecretAccessKey, f.Region)
	}
	ui.Info("To manage AWS account ID: %s in region %s", f.AccountID, f.Region)
	ui.Info("Account name in Mantil is %s", f.AccountName)
}

type awsFlags struct {
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

func (f *awsFlags) ParseArgs(args []string) {
	if len(args) == 0 {
		f.AccountName = workspace.DefaultAccountName
		return
	}
	f.AccountName = args[0]
}

func (f *awsFlags) Validate() error {
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

func (f *awsFlags) validateAccessKeys() error {
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

func (f *awsFlags) readFromEnv() error {
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

func (f *awsFlags) awsConnect() (*aws.AWS, error) {
	cli, err := f.awsClient()
	if err != nil {
		return nil, err
	}
	f.AccountID, err = cli.AccountID()
	if err != nil {
		return nil, err
	}
	f.Region = cli.Region()
	return cli, nil
}

func (f *awsFlags) awsClient() (*aws.AWS, error) {
	if f.Profile != "" {
		return aws.NewFromProfile(f.Profile)
	}
	return aws.NewWithCredentials(f.AccessKeyID, f.SecretAccessKey, "", f.Region)
}
