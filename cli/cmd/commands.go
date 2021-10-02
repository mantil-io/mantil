package cmd

import (
	"fmt"
	"time"

	"github.com/mantil-io/mantil/cli/cmd/project"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
	"github.com/spf13/cobra"
)

func newDestroyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy all infrastructure resources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initDestroy(cmd, args).run()
		},
	}
	cmd.Flags().Bool("repo", false, "delete local repository")
	cmd.Flags().Bool("force", false, "don't ask for confirmation")
	cmd.Flags().StringP("stage", "s", "", "name of the stage to destroy, if left empty all stages will be destroyed")
	return cmd
}

func newEnvCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Show project environment variables",
		Long: `Show project environment variables

You can set environment variables in terminal with:
$ eval $(mantil env)
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initEnv(cmd, args).run()
		},
	}
	cmd.Flags().BoolP("url", "u", false, "show only project api url")
	cmd.Flags().StringP("stage", "s", "", "stage name")
	return cmd
}

func newInvokeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoke <function>[/method]",
		Short: "Makes requests to functions through project's API Gateway",
		Args:  cobra.ExactArgs(1),
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

For the description of filter patterns see:
https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return initLogs(cmd, args).run()
		},
	}
	cmd.Flags().StringP("filter-pattern", "p", "", "filter pattern to use")
	cmd.Flags().DurationP("since", "s", 3*time.Hour, "from what time to begin displaying logs, default is 3 hours ago")
	cmd.Flags().BoolP("tail", "f", false, "continuously poll for new logs")
	cmd.Flags().String("stage", "", "name of the stage to fetch logs for")
	return cmd
}

func newNewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new <project>",
		Short: "Initializes a new Mantil project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return initNew(cmd, args).run()
		},
	}
	cmd.Flags().String("from", "", "name of the template or URL of the repository that will be used as one")
	cmd.Flags().String("module-name", "", "replace module name and import paths")
	return cmd
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
		Short: "Watch for file changes and automatically deploy functions",
		Args:  cobra.NoArgs,
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

func initDestroy(cmd *cobra.Command, args []string) *destroyCmd {
	force, _ := cmd.Flags().GetBool("force")
	stageName, _ := cmd.Flags().GetString("stage")
	deleteRepo, _ := cmd.Flags().GetBool("repo")

	ctx := project.MustContext()
	if !force {
		confirmProjectDestroy(ctx.Project, stageName)
	}

	return &destroyCmd{
		stageName:  stageName,
		ctx:        ctx,
		deleteRepo: deleteRepo,
	}
}

func initEnv(cmd *cobra.Command, args []string) *envCmd {
	url, _ := cmd.Flags().GetBool("url")
	stageName, _ := cmd.Flags().GetString("stage")

	return &envCmd{
		url:       url,
		stageName: stageName,
	}
}

func initInvoke(cmd *cobra.Command, args []string) *invokeCmd {
	stageName, _ := cmd.Flags().GetString("stage")
	data := cmd.Flag("data").Value.String()
	includeHeaders, _ := cmd.Flags().GetBool("include")
	includeLogs, _ := cmd.Flags().GetBool("logs")

	ctx := project.MustContextWithStage(stageName)

	return &invokeCmd{
		endpoint:       args[0],
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
	p, path := getProject()
	run := cmd.Flag("run").Value.String()
	stageName, _ := cmd.Flags().GetString("stage")

	return &testCmd{
		project:   p,
		stageName: stageName,
		repoPath:  path,
		runRegexp: run,
	}
}

func getProject() (*workspace.Project, string) {
	path, err := workspace.FindProjectRoot(".")
	if err != nil {
		ui.Fatal(err)
	}
	p, err := workspace.LoadProject(path)
	if err != nil {
		ui.Error(err)
	}
	return p, path
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
		endpoint:       method,
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

func confirmProjectDestroy(p *workspace.Project, stageName string) {
	var label string
	if stageName == "" {
		label = "To confirm deletion of all stages, please enter the project name"
	} else {
		label = fmt.Sprintf("To confirm deletion of stage %s, please enter the project name", stageName)
	}
	confirmationPrompt := promptui.Prompt{
		Label: label,
	}
	projectName, err := confirmationPrompt.Run()
	if err != nil {
		ui.Fatal(err)
	}
	if p.Name != projectName {
		ui.Fatalf("Project name doesn't match, exiting...")
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
