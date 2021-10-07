package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/cmd/generate"
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
			ui.Info("==> Next steps:")
			ui.Info("\t- Run mantil help to get started")
			ui.Info("\t- Run mantil new to start a new project")
			ui.Info("\t- Further documentation:")
			ui.Info("\t  https://docs.mantil.io")
			ui.Info("") // new line
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

func newEnvCommand() *cobra.Command {
	f := &envFlags{}
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Show project environment variables",
		Long: `Show project environment variables

You can set environment variables in terminal with:
$ eval $(mantil env)
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			e, err := newEnv(f)
			if err != nil {
				return log.Wrap(err)
			}
			if err := e.run(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&f.url, "url", "u", false, "show only project api url")
	cmd.Flags().StringVarP(&f.stage, "stage", "s", "", "stage name")
	return cmd
}

func newInvokeCommand() *cobra.Command {
	f := &invokeFlags{}
	cmd := &cobra.Command{
		Use:   "invoke <function>[/method]",
		Short: "Invoke function methods through the project's API Gateway",
		Long: `Invoke function methods through the project's API Gateway

This is a convenience method and provides similar output to calling:
curl -X POST https://<stage_api_url>/<function>[/method] [-d '<data>'] [-I]

Additionally, you can enable streaming of lambda execution logs by setting the --logs flag.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.path = args[0]
			i, err := newInvoke(f)
			if err != nil {
				return log.Wrap(err)
			}
			if err := i.run(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.data, "data", "d", "", "data for the method invoke request")
	cmd.Flags().BoolVarP(&f.includeHeaders, "include", "i", false, "include response headers in the output")
	cmd.Flags().BoolVarP(&f.includeLogs, "logs", "l", false, "show lambda execution logs")
	cmd.Flags().StringVarP(&f.stage, "stage", "s", "", "name of the stage to target")
	return cmd
}

func newLogsCommand() *cobra.Command {
	f := &logsFlags{}
	cmd := &cobra.Command{
		Use:   "logs [function]",
		Short: "Fetch logs for a specific function/api",
		Long: `Fetch logs for a specific function/api

Logs can be filtered using Cloudwatch filter patterns. For more information see:
https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html

If the --tail flag is set the process will keep running and polling for new logs every second.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				f.function = args[0]
			}
			l, err := newLogs(f)
			if err != nil {
				return log.Wrap(err)
			}
			if err := l.run(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.filter, "filter-pattern", "p", "", "filter pattern to use")
	cmd.Flags().DurationVarP(&f.since, "since", "s", 3*time.Hour, "from what time to begin displaying logs, default is 3 hours ago")
	cmd.Flags().BoolVarP(&f.tail, "tail", "t", false, "continuously poll for new logs")
	cmd.Flags().StringVar(&f.stage, "stage", "", "name of the stage to fetch logs for")
	return cmd
}

func newNewCommand() *cobra.Command {
	f := &newFlags{}
	cmd := &cobra.Command{
		Use:   "new <project>",
		Short: "Initializes a new Mantil project",
		Long: fmt.Sprintf(`Initializes a new Mantil project

This command will initialize a new Mantil project from the source provided with the --from flag.
The source can either be an existing git repository or one of the predefined templates:
%s

If no source is provided it will default to the template "%s".

By default, the go module name of the initialized project will be the project name.
This can be changed by setting the --module-name flag.`, templateList(), defaultTemplate),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.name = args[0]
			n, err := newNew(f)
			if err != nil {
				return log.Wrap(err)
			}
			if err := n.run(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.repo, "from", "", "name of the template or URL of the repository that will be used as one")
	cmd.Flags().StringVar(&f.moduleName, "module-name", "", "replace module name and import paths")
	return cmd
}

func templateList() string {
	var items []string
	for t, r := range templateRepos {
		items = append(items, fmt.Sprintf("%s - %s", t, r))
	}
	return strings.Join(items, "\n")
}

func newTestCommand() *cobra.Command {
	f := &testFlags{}
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run project integration tests",
		Long: `Run project integration tests

Project integration tests are pure Go test in [project-root]/test folder.
Mantil sets MANTIL_API_URL environment variable to point to the current
project api url and runs tests with 'go test -v'.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := newTest(f)
			if err != nil {
				return log.Wrap(err)
			}
			if err := t.run(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.runRegexp, "run", "r", "", "run only tests with this pattern in name")
	cmd.Flags().StringVarP(&f.stage, "stage", "s", "", "stage name")
	return cmd
}

func newWatchCommand() *cobra.Command {
	f := &watchFlags{}
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch for file changes and automatically deploy them",
		Long: `Watch for file changes and automatically deploy them

This command will start a watcher process that listens to changes in any .go files in the project directory
and automatically deploys changes to the stage provided via the --stage flag.

Optionally, you can set a method to invoke after every deploy using the --method, --data and --test flags.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w, err := newWatch(f)
			if err != nil {
				return log.Wrap(err)
			}
			if err := w.run(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.method, "method", "m", "", "method to invoke after deploying changes")
	cmd.Flags().StringVarP(&f.data, "data", "d", "", "data for the method invoke request")
	cmd.Flags().BoolVarP(&f.test, "test", "t", false, "run tests after deploying changes")
	cmd.Flags().StringVarP(&f.stage, "stage", "s", "", "name of the stage to deploy changes to")
	return cmd
}

func newStageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stage",
		Short: "Manage project stages",
		Long: `Manage project stages

A stage represents a named deployment of the project. Each stage creates a set of resources
which can be managed and configured separately.

Stages can be deployed to any account in the workspace.`,
	}
	cmd.AddCommand(newStageNewCommand())
	cmd.AddCommand(newStageDestroyCommand())
	return cmd
}

func newStageNewCommand() *cobra.Command {
	f := &stageFlags{}
	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new stage",
		Long: fmt.Sprintf(`Create a new stage

This command will create a new stage with the given name. If the name is left empty it will default to "%s".

If only one account is set up in the workspace, the stage will be deployed to that account by default.
Otherwise, you will be asked to pick an account. The account can also be specified via the --account flag.`, workspace.DefaultStageName),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				f.stage = args[0]
			}
			s, err := newStage(f)
			if err != nil {
				return log.Wrap(err)
			}
			if err := s.new(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.account, "account", "a", "", "account in which the stage will be created")
	return cmd
}

func newStageDestroyCommand() *cobra.Command {
	f := &stageFlags{}
	cmd := &cobra.Command{
		Use:   "destroy <name>",
		Short: "Destroy a stage",
		Long: `Destroy a stage

This command will destroy all resources belonging to a stage.
Optionally, you can set the --all flag to destroy all stages.

By default you will be asked to confirm the destruction by typing in the project name.
This behavior can be disabled using the --force flag.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				f.stage = args[0]
			}
			s, err := newStage(f)
			if err != nil {
				log.Wrap(err)
			}
			if err := s.destroy(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&f.force, "force", false, "don't ask for confirmation")
	cmd.Flags().BoolVar(&f.destroyAll, "all", false, "destroy all stages")
	return cmd
}

func newGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Automatically generate code in the project",
	}
	cmd.AddCommand(newGenerateApiCommand())
	return cmd
}

func newGenerateApiCommand() *cobra.Command {
	f := &generate.Flags{}
	cmd := &cobra.Command{
		Use:   "api <function>",
		Short: "Generate Go code for a new API",
		Long: `Generate Go code for new API

This command generates all the boilerplate code necessary to get started writing a new API.
An API is a lambda function with at least one (default) request/response method.

Optionally, you can define additional methods using the --methods flag. Each method will have a separate
entrypoint and request/response structures.

After being deployed the can then be invoked using mantil invoke, for example:

mantil invoke ping
mantil invoke ping/hello`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f.Name = args[0]
			g, err := generate.New(f)
			if err != nil {
				return log.Wrap(err)
			}
			if err := g.Api(); err != nil {
				return log.Wrap(err)
			}
			ui.Notice("successfuly generated api %s", f.Name)
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&f.Methods, "methods", "m", nil, "additional function methods, if left empty only the Default method will be created")
	return cmd
}

func newDeployCommand() *cobra.Command {
	f := &deploy.Flags{}
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploys updates to stages",
		Long: `Deploys updates to stages

This command checks if any assets, code or configuration have changed since the last deployment
and applies the necessary updates.

The --stage flag accepts any existing stage and defaults to the default stage if omitted.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := deploy.New(f)
			if err != nil {
				return log.Wrap(err)
			}
			_, err = d.Deploy()
			if err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.Stage, "stage", "s", "", "the name of the stage to deploy to")
	return cmd
}
