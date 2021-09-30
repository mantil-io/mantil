package cmd

import (
	"fmt"
	"os"

	"github.com/mantil-io/mantil/workspace"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/cmd/setup"
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
			cred, err := initAwsInstall(cmd, args)
			if err != nil {
				return err
			}
			if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
				showDryRunInfo(cred)
				return nil
			}
			return cred.cmd.Create()
		},
	}
	addAwsInstallFlags(cmd)
	cmd.Flags().Bool("override", false, "force override access tokens on already installed account")
	return cmd
}

func newAwsUninstallCommand() *cobra.Command {
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
			cred, err := initAwsInstall(cmd, args)
			if err != nil {
				return err
			}
			if dryRun, _ := cmd.Flags().GetBool("dry-run"); dryRun {
				showDryRunInfo(cred)
				return nil
			}

			return cred.cmd.Destroy()
		},
	}
	addAwsInstallFlags(cmd)
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

func addAwsInstallFlags(cmd *cobra.Command) {
	cmd.Flags().String("aws-access-key-id", "", "access key ID for the AWS account, must be used with the aws-secret-access-key and aws-region flags")
	cmd.Flags().String("aws-secret-access-key", "", "secret access key for the AWS account, must be used with the aws-access-key-id and aws-region flags")
	cmd.Flags().String("aws-region", "", "region for the AWS account, must be used with and aws-access-key-id and aws-secret-access-key flags")
	cmd.Flags().Bool("aws-env", false, "use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION environment variables for AWS authentication")
	cmd.Flags().String("aws-profile", "", "use the given profile for AWS authentication")
	cmd.Flags().Bool("dry-run", false, "don't start install/uninstall just show what credentials will be used")
}

func showDryRunInfo(cred *credentials) {
	accountID, _ := cred.cli.AccountID()
	if cred.profile != "" {
		fmt.Printf(`command will use AWS profile %s defined in you AWS configuration file (~/.aws/config)
to manage AWS account ID: %s in region %s
`, cred.profile, accountID, cred.cli.Region())
	} else {
		fmt.Printf(`command will use AWS credentials:
  aws-access-key-id: %s
  aws-secret-access-key: %s
  aws-region: %s
to manage AWS account ID: %s in region %s
`, cred.accessKeyID, cred.secretAccessKey, cred.region, accountID, cred.cli.Region())
	}
	fmt.Printf("account name in Mantil is %s\n", cred.accountName)
}

type credentials struct {
	accessKeyID     string
	secretAccessKey string
	sessionToken    string
	region          string
	profile         string
	accountName     string
	cli             *aws.AWS
	version         *setup.VersionInfo
	cmd             *setup.Cmd
}

func initAwsInstall(cmd *cobra.Command, args []string) (*credentials, error) {
	c := credentials{
		accountName: workspace.DefaultAccountName,
	}
	if len(args) > 0 {
		c.accountName = args[0]
	}

	if err := c.createAwsClient(cmd); err != nil {
		return nil, err
	}

	v, ok := setup.GetVersion(cmd.Context())
	if !ok {
		return nil, fmt.Errorf("version not found in context")
	}
	c.version = v

	if err := c.cli.Try(); err != nil {
		return nil, err
	}

	override, _ := cmd.Flags().GetBool("override")
	c.cmd = setup.New(c.cli, c.version, c.accountName, override)
	return &c, nil
}

func (c *credentials) createAwsClient(cmd *cobra.Command) error {
	if err := c.read(cmd); err != nil {
		return err
	}
	if c.profile != "" {
		cli, err := aws.NewFromProfile(c.profile)
		if err != nil {
			return err
		}
		c.cli = cli
		return nil
	}
	cli, err := aws.NewWithCredentials(c.accessKeyID, c.secretAccessKey, "", c.region)
	if err != nil {
		return err
	}
	c.cli = cli
	return nil
}

func (c *credentials) read(cmd *cobra.Command) error {
	var err error
	if cmd.Flags().Changed("aws-access-key-id") {
		err = c.awsFromAccessKeys(cmd)
	} else if cmd.Flags().Changed("aws-env") {
		err = c.awsFromEnv(cmd)
	} else if cmd.Flags().Changed("aws-profile") {
		err = c.awsFromProfile(cmd)
	} else {
		err = fmt.Errorf("aws credentials not provided")
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *credentials) awsFromAccessKeys(cmd *cobra.Command) error {
	var err error
	c.accessKeyID, err = cmd.Flags().GetString("aws-access-key-id")
	if err != nil {
		return err
	}
	if c.accessKeyID == "" {
		return fmt.Errorf("access key id not provided, must be used with the aws-secret-access-key and aws-region")
	}
	c.secretAccessKey, err = cmd.Flags().GetString("aws-secret-access-key")
	if err != nil {
		return err
	}
	if c.secretAccessKey == "" {
		return fmt.Errorf("secret access key not provided, must be used with the aws-access-key-id and aws-region flags")
	}
	c.region, err = cmd.Flags().GetString("aws-region")
	if err != nil {
		return err
	}
	if c.region == "" {
		return fmt.Errorf("region not provided, must be used with and aws-access-key-id and aws-secret-access-key flags")
	}
	return nil
}

func (c *credentials) awsFromEnv(cmd *cobra.Command) error {
	var ok bool
	c.accessKeyID, ok = os.LookupEnv("AWS_ACCESS_KEY_ID")
	if !ok || c.accessKeyID == "" {
		return fmt.Errorf("access key id not provided")
	}
	c.secretAccessKey, ok = os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !ok || c.secretAccessKey == "" {
		return fmt.Errorf("secret access key not provided")
	}
	c.region, ok = os.LookupEnv("AWS_DEFAULT_REGION")
	if !ok || c.region == "" {
		return fmt.Errorf("region not provided")
	}
	return nil
}

func (c *credentials) awsFromProfile(cmd *cobra.Command) error {
	var err error
	c.profile, err = cmd.Flags().GetString("aws-profile")
	if err != nil {
		return err
	}
	if c.profile == "" {
		return fmt.Errorf("profile not provided")
	}
	return nil
}
