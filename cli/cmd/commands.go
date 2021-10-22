package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/cmd/generate"
	"github.com/mantil-io/mantil/cli/controller"
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
	a := &controller.SetupArgs{}
	cmd := &cobra.Command{
		Use: boldize(fmt.Sprintf(`install [account-name] [flags]

\bARGUMENTS\c
  [account-name]  Mantil account name reference.
                  If not provided default name %s will be used for the first account.`,
			controller.DefaultAccountName())),
		Short: "Install Mantil into AWS account",
		Long: `Install Mantil into AWS account

Command will install backend services into AWS account.
You must provide credentials for Mantil to access your AWS account.

There is --dry-run flag which will show you what credentials will be used
and what account will be managed by command.`,
		Args:    cobra.MaximumNArgs(1),
		Example: credentialsHelp("install"),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.ParseArgs(args)
			stp, err := controller.NewSetup(a)
			if err != nil {
				return log.Wrap(err)
			}
			if a.DryRun {
				showAwsDryRunInfo(a)
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
	bindAwsInstallFlags(cmd, a)
	cmd.Flags().BoolVar(&a.Override, "override", false, "force override access tokens on already installed account")
	return cmd
}

func newAwsUninstallCommand() *cobra.Command {
	a := &controller.SetupArgs{}
	cmd := &cobra.Command{
		Use: boldize(fmt.Sprintf(`uninstall [account-name] [flags]

\bARGUMENTS\c
  [account-name]  Mantil account name reference.
                  If not provided default name %s will be used for the first account.`, workspace.DefaultAccountName)),
		Short: "Uninstall Mantil from AWS account",
		Long: `Uninstall Mantil from AWS account

Command will remove backend services from AWS account.
You must provide credentials for Mantil to access your AWS account.

There is --dry-run flag which will show you what credentials will be used
and what account will be managed by command.`,
		Args:    cobra.MaximumNArgs(1),
		Example: credentialsHelp("uninstall"),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.ParseArgs(args)
			stp, err := controller.NewSetup(a)
			if err != nil {
				return log.Wrap(err)
			}
			if a.DryRun {
				showAwsDryRunInfo(a)
				return nil
			}
			return stp.Destroy()
		},
	}
	bindAwsInstallFlags(cmd, a)
	return cmd
}

func credentialsHelp(commandName string) string {
	return strings.ReplaceAll(`  You must provide credentials for Mantil to access your AWS account.
  There are three ways to provide credentials.

  ==> specifiy access keys as arguments:
  $ mantil aws {.CommandName} --aws-access-key-id=AKIAIOSFODNN7EXAMPLE \
                       --aws-secret-access-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
                       --aws-region=us-east-1

  ==> read access keys from environment variables:
  $ export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
  $ export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
  $ export AWS_DEFAULT_REGION=us-east-1
  $ mantil aws {.CommandName} --aws-env

  Reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html

  ==> use your named AWS profile form ~/.aws/config
  $ mantil aws {.CommandName} --aws-profile=my-named-profile

  Reference: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html`, "{.CommandName}", commandName)
}

func bindAwsInstallFlags(cmd *cobra.Command, a *controller.SetupArgs) {
	cmd.Flags().StringVar(&a.AccessKeyID, "aws-access-key-id", "", "access key ID for the AWS account, must be used with the aws-secret-access-key and aws-region flags")
	cmd.Flags().StringVar(&a.SecretAccessKey, "aws-secret-access-key", "", "secret access key for the AWS account, must be used with the aws-access-key-id and aws-region flags")
	cmd.Flags().StringVar(&a.Region, "aws-region", "", "region for the AWS account, must be used with and aws-access-key-id and aws-secret-access-key flags")
	cmd.Flags().BoolVar(&a.UseEnv, "aws-env", false, "use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION environment variables for AWS authentication")
	cmd.Flags().StringVar(&a.Profile, "aws-profile", "", "use the given profile for AWS authentication")
	cmd.Flags().BoolVar(&a.DryRun, "dry-run", false, "don't start install/uninstall just show what credentials will be used")
}

func showAwsDryRunInfo(a *controller.SetupArgs) {
	if a.Profile != "" {
		ui.Info(`Command will use AWS profile %s defined in your AWS configuration file (~/.aws/config)`, a.Profile)
	} else {
		ui.Info(`Command will use AWS credentials:
  aws-access-key-id: %s
  aws-secret-access-key: %s
  aws-region: %s`, a.AccessKeyID, a.SecretAccessKey, a.Region)
	}
	ui.Info("To manage AWS account ID: %s in region %s", a.AccountID, a.Region)
	ui.Info("Account name in Mantil is %s", a.AccountName)
}

func newEnvCommand() *cobra.Command {
	var a controller.EnvArgs
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Show project environment variables",
		Long: `Show project environment variables

You can set environment variables in terminal with:
$ eval $(mantil env)
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := controller.Env(a)
			if err == nil {
				fmt.Printf("%s", out)
			}
			return err
		},
	}
	cmd.Flags().BoolVarP(&a.Url, "url", "u", false, "show only project api url")
	cmd.Flags().StringVarP(&a.Stage, "stage", "s", "", "stage name")
	return cmd
}

func newInvokeCommand() *cobra.Command {
	var a controller.InvokeArgs
	cmd := &cobra.Command{
		Use:   "invoke <function>[/method]",
		Short: "Invoke function methods through the project's API Gateway",
		Long: `Invoke function methods through the project's API Gateway

This is a convenience method and provides similar output to calling:
curl -X POST https://<stage_api_url>/<function>[/method] [-d '<data>'] [-I]

Additionally, you can enable streaming of lambda execution logs by setting the --logs flag.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Path = args[0]
			if err := controller.Invoke(a); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.Data, "data", "d", "", "data for the method invoke request")
	cmd.Flags().BoolVarP(&a.IncludeHeaders, "include", "i", false, "include response headers in the output")
	cmd.Flags().BoolVarP(&a.IncludeLogs, "logs", "l", false, "show lambda execution logs")
	cmd.Flags().StringVarP(&a.Stage, "stage", "s", "", "name of the stage to target")
	return cmd
}

func newLogsCommand() *cobra.Command {
	var a logsArgs
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
				a.function = args[0]
			}
			l, err := newLogs(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := l.run(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.filter, "filter-pattern", "p", "", "filter pattern to use")
	cmd.Flags().DurationVarP(&a.since, "since", "s", 3*time.Hour, "from what time to begin displaying logs, default is 3 hours ago")
	cmd.Flags().BoolVarP(&a.tail, "tail", "t", false, "continuously poll for new logs")
	cmd.Flags().StringVar(&a.stage, "stage", "", "name of the stage to fetch logs for")
	return cmd
}

func newNewCommand() *cobra.Command {
	var a newArgs
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
			a.name = args[0]
			n, err := newNew(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := n.run(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&a.repo, "from", "", "name of the template or URL of the repository that will be used as one")
	cmd.Flags().StringVar(&a.moduleName, "module-name", "", "replace module name and import paths")
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
	var a testArgs
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
			t, err := newTest(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := t.run(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.runRegexp, "run", "r", "", "run only tests with this pattern in name")
	cmd.Flags().StringVarP(&a.stage, "stage", "s", "", "stage name")
	return cmd
}

func newWatchCommand() *cobra.Command {
	var a watchArgs
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch for file changes and automatically deploy them",
		Long: `Watch for file changes and automatically deploy them

This command will start a watcher process that listens to changes in any .go files in the project directory
and automatically deploys changes to the stage provided via the --stage flag.

Optionally, you can set a method to invoke after every deploy using the --method, --data and --test flags.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			w, err := newWatch(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := w.run(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.method, "method", "m", "", "method to invoke after deploying changes")
	cmd.Flags().StringVarP(&a.data, "data", "d", "", "data for the method invoke request")
	cmd.Flags().BoolVarP(&a.test, "test", "t", false, "run tests after deploying changes")
	cmd.Flags().StringVarP(&a.stage, "stage", "s", "", "name of the stage to deploy changes to")
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
	var a stageArgs
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
				a.stage = args[0]
			}
			s, err := newStage(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := s.new(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.account, "account", "a", "", "account in which the stage will be created")
	return cmd
}

func newStageDestroyCommand() *cobra.Command {
	var a stageArgs
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
				a.stage = args[0]
			}
			s, err := newStage(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := s.destroy(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&a.force, "force", false, "don't ask for confirmation")
	cmd.Flags().BoolVar(&a.destroyAll, "all", false, "destroy all stages")
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
	var methods []string
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
			msg := fmt.Sprintf("Generating function %s", args[0])
			if len(methods) > 0 {
				msg = fmt.Sprintf("%s with additional methods %s", msg, strings.Join(methods, ","))
			}
			ui.Info("%s.\n", msg)
			if err := generate.Api(args[0], methods); err != nil {
				return log.Wrap(err)
			}
			ui.Notice("\nDone.")
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&methods, "methods", "m", nil, "additional function methods, if left empty only the Default method will be created")
	return cmd
}

func newDeployCommand() *cobra.Command {
	var a deploy.Args
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploys updates to stages",
		Long: `Deploys updates to stages

This command checks if any assets, code or configuration have changed since the last deployment
and applies the necessary updates.

The --stage flag accepts any existing stage and defaults to the default stage if omitted.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := deploy.New(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := d.Deploy(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.Stage, "stage", "s", "", "the name of the stage to deploy to")
	return cmd
}
