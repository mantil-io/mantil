package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/mantil-io/mantil/cli/cmd/project"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
	"github.com/spf13/cobra"
)

func newEnvCommand() *cobra.Command {
	var f envCmd
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Show project environment variables",
		Long: `Show project environment variables

You can set environment variables in terminal with:
$ eval $(mantil env)
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return f.run()
		},
	}
	cmd.Flags().BoolVarP(&f.url, "url", "u", false, "show only project api url")
	cmd.Flags().StringVarP(&f.stage, "stage", "s", "", "stage name")
	return cmd
}

func newInvokeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoke <function>[/method]",
		Short: "Invoke function methods through the project's API Gateway",
		Long: `Invoke function methods through the project's API Gateway

This is a convenience method and provides similar output to calling:
curl -X POST https://<stage_api_url>/<function>[/method] [-d '<data>'] [-I]

Additionally, you can enable streaming of lambda execution logs by setting the --logs flag.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return initInvoke(cmd, args).run()
		},
	}
	cmd.Flags().StringP("data", "d", "", "data for the method invoke request")
	cmd.Flags().BoolP("include", "i", false, "include response headers in the output")
	cmd.Flags().BoolP("logs", "l", false, "show lambda execution logs")
	cmd.Flags().StringP("stage", "s", "", "name of the stage to target")
	return cmd
}

func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs [function]",
		Short: "Fetch logs for a specific function/api",
		Long: `Fetch logs for a specific function/api

Logs can be filtered using Cloudwatch filter patterns. For more information see:
https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html

If the --tail flag is set the process will keep running and polling for new logs every second.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return initLogs(cmd, args).run()
		},
	}
	cmd.Flags().StringP("filter-pattern", "p", "", "filter pattern to use")
	cmd.Flags().DurationP("since", "s", 3*time.Hour, "from what time to begin displaying logs, default is 3 hours ago")
	cmd.Flags().BoolP("tail", "t", false, "continuously poll for new logs")
	cmd.Flags().String("stage", "", "name of the stage to fetch logs for")
	return cmd
}

func newNewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new <project>",
		Short: "Initializes a new Mantil project",
		Long: fmt.Sprintf(`Initializes a new Mantil project

This command will initialize a new Mantil project from the source provided with the --from flag.
The source can either be an existing git repository or one of the predefined templates:
%s

If no source is provided it will default to the template "%s".

By default, the go module name of the initialized project will be the same as in the source.
This can be changed by setting the --module-name flag.`, templateList(), defaultTemplate),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return initNew(cmd, args).run()
		},
	}
	cmd.Flags().String("from", "", "name of the template or URL of the repository that will be used as one")
	cmd.Flags().String("module-name", "", "replace module name and import paths")
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
			return initTest(cmd, args).run()
		},
	}
	cmd.Flags().StringP("run", "r", "", "run only tests with this pattern in name")
	cmd.Flags().StringP("stage", "s", "", "stage name")
	return cmd
}

func newWatchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch for file changes and automatically deploy them",
		Long: `Watch for file changes and automatically deploy them

This command will start a watcher process that listens to changes in any .go files in the project directory
and automatically deploys changes to the stage provided via the --stage flag.

Optionally, you can set a method to invoke after every deploy using the --method, --data and --test flags.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initWatch(cmd, args).run()
		},
	}
	cmd.Flags().BoolP("test", "t", false, "run tests after deploying changes")
	cmd.Flags().StringP("method", "m", "", "method to invoke after deploying changes")
	cmd.Flags().StringP("data", "d", "", "data for the method invoke request")
	cmd.Flags().StringP("stage", "s", "", "name of the stage to deploy changes to")
	return cmd
}

func initInvoke(cmd *cobra.Command, args []string) *invokeCmd {
	stageName, _ := cmd.Flags().GetString("stage")
	data := cmd.Flag("data").Value.String()
	includeHeaders, _ := cmd.Flags().GetBool("include")
	includeLogs, _ := cmd.Flags().GetBool("logs")

	ctx := project.MustContextWithStage(stageName)

	return &invokeCmd{
		path:           args[0],
		ctx:            ctx,
		data:           data,
		includeHeaders: includeHeaders,
		includeLogs:    includeLogs,
	}
}

func initLogs(cmd *cobra.Command, args []string) *logsCmd {
	stageName, _ := cmd.Flags().GetString("stage")
	filter := cmd.Flag("filter-pattern").Value.String()
	since, _ := cmd.Flags().GetDuration("since")
	tail, _ := cmd.Flags().GetBool("tail")

	ctx := project.MustContextWithStage(stageName)
	awsClient := ctx.MustInitialiseAWSSDK()

	var function string
	if len(args) > 0 {
		function = args[0]
	} else {
		function = selectFunctionFromStage(ctx.Stage)
	}
	startTime := time.Now().Add(-since)

	return &logsCmd{
		ctx:       ctx,
		function:  function,
		awsClient: awsClient,
		filter:    filter,
		startTime: startTime,
		tail:      tail,
	}
}

func initNew(cmd *cobra.Command, args []string) *newCmd {
	projectName := args[0]
	repo := cmd.Flag("from").Value.String()
	moduleName := cmd.Flag("module-name").Value.String()
	if moduleName == "" {
		moduleName = projectName
	}

	return &newCmd{
		name:       projectName,
		repo:       repo,
		moduleName: moduleName,
	}
}

func initTest(cmd *cobra.Command, args []string) *testCmd {
	run := cmd.Flag("run").Value.String()
	stageName, _ := cmd.Flags().GetString("stage")

	ctx := project.MustContextWithStage(stageName)

	return &testCmd{
		ctx:       ctx,
		runRegexp: run,
	}
}

func initWatch(cmd *cobra.Command, args []string) *watchCmd {
	method := cmd.Flag("method").Value.String()
	test, _ := cmd.Flags().GetBool("test")
	data := cmd.Flag("data").Value.String()
	stageName, _ := cmd.Flags().GetString("stage")

	ctx := project.MustContextWithStage(stageName)
	awsClient := ctx.MustInitialiseAWSSDK()

	deploy, err := deploy.New(ctx, awsClient)
	if err != nil {
		ui.Fatal(err)
	}

	invoke := &invokeCmd{
		path:           method,
		ctx:            ctx,
		data:           data,
		includeHeaders: false,
		includeLogs:    true,
	}

	return &watchCmd{
		ctx:    ctx,
		deploy: deploy,
		invoke: invoke,
		test:   test,
		data:   data,
	}
}

func selectFunctionFromStage(stage *workspace.Stage) string {
	var funcNames []string
	for _, f := range stage.Functions {
		funcNames = append(funcNames, f.Name)
	}
	prompt := promptui.Select{
		Label: "Select a function",
		Items: funcNames,
	}
	_, function, err := prompt.Run()
	if err != nil {
		ui.Fatal(err)
	}
	return function
}
