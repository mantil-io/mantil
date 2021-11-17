package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mantil-io/mantil/cli/controller"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"

	"github.com/spf13/cobra"
)

func newAwsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: "AWS node subcommand",
		Args:  cobra.NoArgs,
	}
	addCommand(cmd, newAwsInstallCommand())
	addCommand(cmd, newAwsUninstallCommand())
	addCommand(cmd, newNodesList())
	return cmd
}

func newNodesList() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nodes",
		Aliases: []string{"ls"},
		Short:   "List Mantil aws nodes",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Nodes()
		},
	}
	return cmd
}

func newAwsInstallCommand() *cobra.Command {
	nextSteps := `
* Run 'mantil new' to start a project from scratch or choose from an existing template.
Check documentation at https://docs.mantil.io for additional inspiration.
`

	argumentsUsage := fmt.Sprintf(`
  [node-name]  Mantil node name reference.
               If not provided default name %s will be used for the first node.`, domain.DefaultNodeName)

	a := &controller.SetupArgs{}
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "install [node-name] [options]",
		Short:   "Install Mantil into AWS account",
		Long: `Install Mantil into AWS account

Command will install backend services into AWS account.
You must provide credentials for Mantil to access your AWS account.

There is --dry-run option which will show you what credentials will be used
and what account will be managed by command.`,
		Args:    cobra.MaximumNArgs(1),
		Example: setupExamples("install"),
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
			if err := stp.Create(domain.Deployment().GetPath); err != nil {
				return log.Wrap(err)
			}
			ui.Info(nextSteps)
			return nil
		},
	}
	setUsageTemplate(cmd, argumentsUsage)
	bindAwsInstallFlags(cmd, a)
	return cmd
}

func newAwsUninstallCommand() *cobra.Command {
	nextSteps := `
* We are sorry to see you go. Help us make Mantil better by letting us know what you didn’t like at support@mantil.com.
`
	a := &controller.SetupArgs{}
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "uninstall [node-name] [options]",
		Short:   "Uninstall Mantil from AWS account",
		Long: `Uninstall Mantil from AWS account

Command will remove backend services from AWS account.
You must provide credentials for Mantil to access your AWS account.

There is --dry-run option which will show you what credentials will be used
and what account will be managed by command.

By default you will be asked to confirm the destruction.
This behaviour can be disabled using the --force option.`,
		Args:    cobra.MaximumNArgs(1),
		Example: setupExamples("uninstall"),
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
			destroyed, err := stp.Destroy()
			if err != nil {
				return log.Wrap(err)
			}
			if destroyed {
				ui.Info(nextSteps)
			}
			return nil
		},
	}
	cmd.SetUsageTemplate(usageTemplate(fmt.Sprintf(`
  [node-name]  Mantil node name reference.
               If not provided default name %s will be used for the first node.`, domain.DefaultNodeName)))
	bindAwsInstallFlags(cmd, a)
	cmd.Flags().BoolVar(&a.Force, "force", false, "Don't ask for confirmation")
	return cmd
}

func setupExamples(commandName string) string {
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
	cmd.Flags().StringVar(&a.AccessKeyID, "aws-access-key-id", "", "Access key ID for the AWS account, must be used with the aws-secret-access-key and aws-region options")
	cmd.Flags().StringVar(&a.SecretAccessKey, "aws-secret-access-key", "", "Secret access key for the AWS account, must be used with the aws-access-key-id and aws-region options")
	cmd.Flags().StringVar(&a.Region, "aws-region", "", "Region for the AWS account, must be used with and aws-access-key-id and aws-secret-access-key options")
	cmd.Flags().BoolVar(&a.UseEnv, "aws-env", false, "Use AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION environment variables for AWS authentication")
	cmd.Flags().StringVar(&a.Profile, "aws-profile", "", "Use the given profile for AWS authentication")
	cmd.Flags().BoolVar(&a.DryRun, "dry-run", false, "Don't start install/uninstall just show what credentials will be used")
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
	ui.Info("Node name in Mantil is %s", a.NodeName)
}

func newEnvCommand() *cobra.Command {
	var a controller.EnvArgs
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Export project environment variables",
		Long: `Export project environment variables
for use in other shell commands.

Mantil project is determined by the current shell folder. It can be anywhere in
the project tree.
If not specified (--stage option) default project stage is used.`,
		Example: `  ==> Set environment variables in terminal
  $ eval $(mantil env)

  ==> Use current stage api url in other shell commands
  $ curl -X POST $(mantil env -url)/ping`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := controller.Env(a)
			if err == nil {
				fmt.Printf("%s", out)
			}
			return err
		},
	}
	cmd.Flags().BoolVarP(&a.Url, "url", "u", false, "Show only project api url")
	cmd.Flags().StringVarP(&a.Stage, "stage", "s", "", "Target project stage")
	return cmd
}

func newInvokeCommand() *cobra.Command {
	var a controller.InvokeArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "invoke <api>[/method]",
		Short:   "Invoke api method for current project and stage",
		Long: `Invoke api method for current project and stage

Makes HTTP request to the gateway endpoint of the project stage. That invokes
lambda function of that project api. If api method is not specified default
(named Default in Go code) is assumed.

Mantil project is determined by the current shell folder. It can be anywhere in
the project tree.
If not specified (--stage option) default project stage is used.

During lambda function execution their logs are shown in terminal. Each lambda
function log line is preffixed with λ symbol. You can hide that logs with the
--no-log option.

This is a convenience method and provides similar output to calling:
$ curl -X POST https://<stage_endpoint_url>/<api>[/method] [-d '<data>'] [-i]`,
		Example: `==> invoke Default method in Ping api
$ mantil invoke ping
200 OK
pong

==> invoke Hello method in Ping api with 'Mantil' data
$ mantil invoke ping/hello -d 'Mantil'
200 OK
Hello, Mantil

==> invoke ReqRsp method in Ping api with json data payload
$ mantil invoke ping/reqrsp -d '{"name":"Mantil"}'
200 OK
{
   "Response": "Hello, Mantil"
}`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Path = args[0]
			if err := controller.Invoke(a); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.Data, "data", "d", "", "Data for the method invoke request")
	cmd.Flags().BoolVarP(&a.IncludeHeaders, "include", "i", false, "Include response headers in the output")
	cmd.Flags().BoolVarP(&a.ExcludeLogs, "no-logs", "n", false, "Hide lambda execution logs")
	cmd.Flags().StringVarP(&a.Stage, "stage", "s", "", "Target project stage")
	return cmd
}

func newLogsCommand() *cobra.Command {
	var a controller.LogsArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "logs <function>",
		Short:   "Fetch logs for a specific function/api",
		Long: `Fetch logs for a specific function/api

Logs can be filtered using Cloudwatch filter patterns. For more information see:
https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html

If the --tail option is set the process will keep running and polling for new logs every second.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Function = args[0]
			if err := controller.Logs(a); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.Filter, "filter-pattern", "p", "", "Filter pattern to use")
	cmd.Flags().DurationVarP(&a.Since, "since", "s", 3*time.Hour, "From what time to begin displaying logs, default is 3 hours ago")
	cmd.Flags().BoolVarP(&a.Tail, "tail", "t", false, "Continuously poll for new logs")
	cmd.Flags().StringVar(&a.Stage, "stage", "", "Name of the stage to fetch logs for")
	return cmd
}

func newNewCommand() *cobra.Command {
	nextSteps := `
* It's time to start developing in the cloud. Run 'mantil stage new' to
create your first development environment or check the documentation at
https://docs.mantil.io for more details.
`
	var a controller.NewArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "new <project>",
		Short:   "Initializes a new Mantil project",
		Long: fmt.Sprintf(`Initializes a new Mantil project

This command will initialize a new Mantil project from the source provided with the --from option.
The source can either be an existing git repository or one of the predefined templates:
%s

If no source is provided it will default to the template "%s".

By default, the go module name of the initialized project will be the project name.
This can be changed by setting the --module-name option.`, templateList(), controller.DefaultTemplate),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Name = args[0]
			if err := controller.New(a); err != nil {
				return log.Wrap(err)
			}
			ui.Info(nextSteps)
			return nil
		},
	}
	cmd.Flags().StringVar(&a.From, "from", "", "Name of the template or URL of the repository that will be used as one")
	cmd.Flags().StringVar(&a.ModuleName, "module-name", "", "Replace module name and import paths")
	return cmd
}

func templateList() string {
	var items []string
	for t, r := range controller.TemplateRepos {
		items = append(items, fmt.Sprintf("%s - %s", t, r))
	}
	return strings.Join(items, "\n")
}

func newTestCommand() *cobra.Command {
	var a controller.TestArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "test",
		Short:   "Run project integration tests",
		Long: `Run project integration tests

Project integration tests are pure Go test in [project-root]/test folder.
Mantil sets MANTIL_API_URL environment variable to point to the current
project api url and runs tests with 'go test -v'.
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := controller.Test(a)
			if err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.RunRegexp, "run", "r", "", "Run only tests with this pattern in name")
	cmd.Flags().StringVarP(&a.Stage, "stage", "s", "", "Stage name")
	return cmd
}

func newWatchCommand() *cobra.Command {
	var a controller.WatchArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "watch",
		Short:   "Watch for file changes and automatically deploy them",
		Long: `Watch for file changes and automatically deploy them

This command will start a watcher process that listens to changes in any .go files in the project directory
and automatically deploys changes to the stage provided via the --stage option.

Optionally, you can set a method to invoke after every deploy using the --method, --data and --test options.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := controller.Watch(a); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.Method, "method", "m", "", "Method to invoke after deploying changes")
	cmd.Flags().StringVarP(&a.Data, "data", "d", "", "Data for the method invoke request")
	cmd.Flags().StringVarP(&a.Stage, "stage", "s", "", "Name of the stage to deploy changes to")
	cmd.Flags().BoolVarP(&a.Test, "test", "t", false, "Run tests after deploying changes")
	return cmd
}

func newStageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stage",
		Short: "Manage project stages",
		Long: `Manage project stages

A stage represents a named deployment of the project. Each stage creates a set of resources
which can be managed and configured separately.

Stages can be deployed to any node in the workspace.`,
	}
	addCommand(cmd, newStageNewCommand())
	addCommand(cmd, newStageDestroyCommand())
	addCommand(cmd, newStageList())
	addCommand(cmd, newStageUse())
	return cmd
}

func newStageNewCommand() *cobra.Command {
	nextSteps := `
* Try 'mantil invoke' to see your fully functional Mantil serverless application in action.
`
	var a controller.StageArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "new <name>",
		Short:   "Create a new stage",
		Long: fmt.Sprintf(`Create a new stage

This command will create a new stage with the given name. If the name is left empty it will default to "%s".

If only one node is set up in the workspace, the stage will be deployed to that node by default.
Otherwise, you will be asked to pick a node. The node can also be specified via the --node option.`, domain.DefaultStageName),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				a.Stage = args[0]
			}
			s, err := controller.NewStage(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := s.New(); err != nil {
				return log.Wrap(err)
			}
			ui.Info(nextSteps)
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.Node, "node", "n", "", "Node in which the stage will be created")
	return cmd
}

func newStageDestroyCommand() *cobra.Command {
	var a controller.StageArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "destroy <name>",
		Short:   "Destroy a stage",
		Long: `Destroy a stage

This command will destroy all resources belonging to a stage.
Optionally, you can set the --all option to destroy all stages.

By default you will be asked to confirm the destruction.
This behavior can be disabled using the --force option.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				a.Stage = args[0]
			}
			s, err := controller.NewStage(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := s.Destroy(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&a.Force, "force", false, "Don't ask for confirmation")
	cmd.Flags().BoolVar(&a.DestroyAll, "all", false, "Destroy all stages")
	return cmd
}

func newStageList() *cobra.Command {
	var a controller.StageArgs
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List stages in project",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := controller.NewStage(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := s.List(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	return cmd
}

func newStageUse() *cobra.Command {
	var a controller.StageArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "use <stage>",
		Short:   "Set default project stage",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Stage = args[0]
			s, err := controller.NewStage(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := s.Use(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	return cmd
}

func newGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Automatically generate code in the project",
	}
	addCommand(cmd, newGenerateApiCommand())
	return cmd
}

func newGenerateApiCommand() *cobra.Command {
	var a controller.GenerateApiArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "api <function>",
		Short:   "Generate Go code for a new API",
		Long: `Generate Go code for new API

This command generates all the boilerplate code necessary to get started writing a new API.
An API is a lambda function with at least one (default) request/response method.

Optionally, you can define additional methods using the --methods option. Each method will have a separate
entrypoint and request/response structures.

After being deployed the can then be invoked using mantil invoke, for example:

mantil invoke ping
mantil invoke ping/hello`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Name = args[0]
			if err := controller.GenerateApi(a); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringSliceVarP(&a.Methods, "methods", "m", nil, "Additional function methods, if left empty only the Default method will be created")
	return cmd
}

func newDeployCommand() *cobra.Command {
	nextSteps := `
* Use 'mantil logs' to see those directly in terminal in an instant.
`
	var a controller.DeployArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "deploy",
		Short:   "Deploys updates to stages",
		Long: `Deploys updates to stages

This command checks if any assets, code or configuration have changed since the last deployment
and applies the necessary updates.

The --stage option accepts any existing stage and defaults to the default stage if omitted.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := controller.NewDeploy(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := d.Deploy(); err != nil {
				return log.Wrap(err)
			}
			ui.Info(nextSteps)
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.Stage, "stage", "s", "", "The name of the stage to deploy to")
	return cmd
}

func newRegisterCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "register",
		Short: "Register Mantil application",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Register()
		},
	}
}

func newActivateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "activate <activation-code>",
		Short: "Activate Mantil application",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Activate(args[0])
		},
	}
}

// this should be used as PreRunE for the commands which needs activation
func ensureActivated(cmd *cobra.Command, args []string) error {
	if !controller.IsActivated() {
		return log.Wrap(log.NotActivatedError)
	}
	return nil
}

func newReportCommand() *cobra.Command {
	var days int
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Make a bug report",
		Long: `Make a bug report

This command sends us your logs so we can analyze them and help you with the issue you're having.

By default last 3 days of logs are included, but you can change that with --days option.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Report(days)
		},
	}
	cmd.Flags().IntVarP(&days, "days", "d", 3, "Days of logs to include in report")
	return cmd
}
