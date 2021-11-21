package cmd

import (
	"fmt"
	"time"

	"github.com/mantil-io/mantil/cli/controller"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/texts"

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
		Short:   "List Mantil AWS nodes",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Nodes()
		},
	}
	return cmd
}

func newAwsInstallCommand() *cobra.Command {
	a := &controller.SetupArgs{}
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "install [node-name] [options]",
		Short:   texts.AwsInstall.Short,
		Long:    texts.AwsInstall.Long,
		Example: texts.AwsInstall.Examples,
		Args:    cobra.MaximumNArgs(1),
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
			showNextSteps(texts.AwsInstall.NextSteps)
			return nil
		},
	}
	setUsageTemplate(cmd, texts.AwsInstall.Arguments)
	bindAwsInstallFlags(cmd, a)
	return cmd
}

func showNextSteps(nextSteps string) {
	if nextSteps == "" {
		return
	}
	ui.Info(nextSteps)
}

func newAwsUninstallCommand() *cobra.Command {
	a := &controller.SetupArgs{}
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "uninstall [node-name] [options]",
		Short:   texts.AwsUninstall.Short,
		Long:    texts.AwsUninstall.Long,
		Example: texts.AwsUninstall.Examples,
		Args:    cobra.MaximumNArgs(1),
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
				showNextSteps(texts.AwsUninstall.NextSteps)
			}
			return nil
		},
	}
	setUsageTemplate(cmd, texts.AwsUninstall.Arguments)
	bindAwsInstallFlags(cmd, a)
	cmd.Flags().BoolVar(&a.Force, "force", false, "Don't ask for confirmation")
	return cmd
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
		Use:     "env",
		Short:   texts.Env.Short,
		Long:    texts.Env.Long,
		Example: texts.Env.Examples,
		Args:    cobra.NoArgs,
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
		Short:   texts.Invoke.Short,
		Long:    texts.Invoke.Long,
		Example: texts.Invoke.Examples,
		Args:    cobra.ExactArgs(1),
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
		Use:     "logs <api-name>",
		Short:   texts.Logs.Short,
		Long:    texts.Logs.Long,
		Args:    cobra.ExactArgs(1),
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
	var a controller.NewArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "new <project>",
		Short:   texts.New.Short,
		Long:    texts.New.Long,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			a.Name = args[0]
			if err := controller.New(a); err != nil {
				return log.Wrap(err)
			}
			showNextSteps(texts.New.NextSteps)
			return nil
		},
	}
	cmd.Flags().StringVar(&a.From, "from", "", "Name of the template or URL of the repository that will be used as one")
	cmd.Flags().StringVar(&a.ModuleName, "module-name", "", "Replace module name and import paths")
	return cmd
}

func newTestCommand() *cobra.Command {
	var a controller.TestArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "test",
		Short:   texts.Test.Short,
		Long:    texts.Test.Long,
		Args:    cobra.NoArgs,
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
		Short:   texts.Watch.Short,
		Long:    texts.Watch.Long,
		Args:    cobra.NoArgs,
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
		Short: texts.Stage.Short,
		Long:  texts.Stage.Long,
	}
	addCommand(cmd, newStageNewCommand())
	addCommand(cmd, newStageDestroyCommand())
	addCommand(cmd, newStageList())
	addCommand(cmd, newStageUse())
	return cmd
}

func newStageNewCommand() *cobra.Command {
	var a controller.StageArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "new <name>",
		Short:   texts.StageNew.Short,
		Long:    texts.StageNew.Long,
		Args:    cobra.MaximumNArgs(1),
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
			showNextSteps(texts.StageNew.NextSteps)
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
		Short:   texts.StageDestroy.Short,
		Long:    texts.StageDestroy.Long,
		Args:    cobra.MaximumNArgs(1),
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
		Short:   texts.StageList.Short,
		Long:    texts.StageList.Long,
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
		Short:   texts.StageUse.Short,
		Long:    texts.StageUse.Long,
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
		Short: texts.Generate.Short,
		Long:  texts.Generate.Long,
	}
	addCommand(cmd, newGenerateApiCommand())
	return cmd
}

func newGenerateApiCommand() *cobra.Command {
	var a controller.GenerateApiArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "api <api-name>",
		Short:   texts.GenerateApi.Short,
		Long:    texts.GenerateApi.Long,
		Args:    cobra.ExactArgs(1),
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
	var a controller.DeployArgs
	cmd := &cobra.Command{
		PreRunE: ensureActivated,
		Use:     "deploy",
		Short:   texts.Deploy.Short,
		Long:    texts.Deploy.Long,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			d, err := controller.NewDeploy(a)
			if err != nil {
				return log.Wrap(err)
			}
			if err := d.Deploy(); err != nil {
				return log.Wrap(err)
			}
			showNextSteps(texts.Deploy.NextSteps)
			return nil
		},
	}
	cmd.Flags().StringVarP(&a.Stage, "stage", "s", "", "The name of the stage to deploy to")
	return cmd
}

func newUserCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: texts.User.Short,
		Long:  texts.User.Long,
		Args:  cobra.NoArgs,
	}
	addCommand(cmd, newRegisterCommand())
	addCommand(cmd, newActivateCommand())
	return cmd
}

func newRegisterCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "register",
		Short: texts.UserRegister.Short,
		Long:  texts.UserRegister.Long,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Register()
		},
	}
}

func newActivateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "activate <activation-token>",
		Short: texts.UserActivate.Short,
		Long:  texts.UserActivate.Long,
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
		Short: texts.Report.Short,
		Long:  texts.Report.Long,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return controller.Report(days)
		},
	}
	cmd.Flags().IntVarP(&days, "days", "d", 3, "Days of logs to include in report")
	return cmd
}
