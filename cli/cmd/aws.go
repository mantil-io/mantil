package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/cmd/setup"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
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
	f := &setup.Flags{}
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
`, credentialsHelp(), setup.DefaultAccountName()),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.ParseArgs(args)
			stp, err := setup.New(f)
			if err != nil {
				return log.Wrap(err)
			}
			if f.DryRun {
				showAwsDryRunInfo(f)
				return nil
			}
			if err := stp.Create(); err != nil {
				return log.WithUserMessage(err, "Install failed!")
			}
			ui.Info(`==> Next steps:
- Run mantil help to get started
- Run mantil new to start a new project
- Further documentation: 
  https://docs.mantil.io`)
			return nil
		},
	}
	bindAwsInstallFlags(cmd, f)
	cmd.Flags().BoolVar(&f.Override, "override", false, "force override access tokens on already installed account")
	return cmd
}

func newAwsUninstallCommand() *cobra.Command {
	f := &setup.Flags{}
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
			stp, err := setup.New(f)
			if err != nil {
				return log.Wrap(err)
			}
			if f.DryRun {
				showAwsDryRunInfo(f)
				return nil
			}
			return stp.Destroy()
		},
	}
	bindAwsInstallFlags(cmd, f)
	return cmd
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

func bindAwsInstallFlags(cmd *cobra.Command, f *setup.Flags) {
	cmd.Flags().StringVar(&f.AccessKeyID, "aws-access-key-id", "", "access key ID for the AWS account, must be used with the aws-secret-access-key and aws-region flags")
	cmd.Flags().StringVar(&f.SecretAccessKey, "aws-secret-access-key", "", "secret access key for the AWS account, must be used with the aws-access-key-id and aws-region flags")
	cmd.Flags().StringVar(&f.Region, "aws-region", "", "region for the AWS account, must be used with and aws-access-key-id and aws-secret-access-key flags")
	cmd.Flags().BoolVar(&f.UseEnv, "aws-env", false, "use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION environment variables for AWS authentication")
	cmd.Flags().StringVar(&f.Profile, "aws-profile", "", "use the given profile for AWS authentication")
	cmd.Flags().BoolVar(&f.DryRun, "dry-run", false, "don't start install/uninstall just show what credentials will be used")
}

func showAwsDryRunInfo(f *setup.Flags) {
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
