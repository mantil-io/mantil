package cmd

import (
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/commands"
	"github.com/mantil-io/mantil/cli/commands/deploy"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/spf13/cobra"
)

func init() {
	addCommandDestroy()
	addCommandEnv()
	addCommandInvoke()
	addCommandLogs()
	addCommandNew()
	addCommandTest()
	addCommandWatch()
}

func addCommandDestroy() {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Destroy all infrastructure resources",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			c := initDestroy(cmd, args)
			if err := c.run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().Bool("repo", false, "delete local repository")
	cmd.Flags().Bool("force", false, "don't ask for confirmation")
	cmd.Flags().StringP("stage", "s", config.DefaultStageName, "stage name")
	rootCmd.AddCommand(cmd)
}

func addCommandEnv() {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Show project environment variables",
		Long: `Show project environment variables

You can set environment variables in terminal with:
$ eval $(mantil env)
`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			c := initEnv(cmd, args)
			if err := c.run(); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().BoolP("url", "u", false, "show only project api url")
	cmd.Flags().StringP("stage", "s", config.DefaultStageName, "stage name")
	rootCmd.AddCommand(cmd)
}

func addCommandInvoke() {
	cmd := &cobra.Command{
		Use:   "invoke <function>[/method]",
		Short: "Makes requests to functions through project's API Gateway",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := initInvoke(cmd, args)
			if err := c.run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringP("data", "d", "", "data for the method invoke request")
	cmd.Flags().BoolP("include", "i", false, "include response headers in the output")
	cmd.Flags().BoolP("logs", "l", false, "show lambda execution logs")
	cmd.Flags().StringP("stage", "s", config.DefaultStageName, "stage name")
	rootCmd.AddCommand(cmd)
}

func addCommandLogs() {
	cmd := &cobra.Command{
		Use:   "logs [function]",
		Short: "Fetch logs for a specific function/api",
		Long: `Fetch logs for a specific function/api

For the description of filter patterns see:
https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := initLogs(cmd, args)
			if err := c.run(); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.Flags().StringP("filter-pattern", "p", "", "filter pattern to use")
	cmd.Flags().DurationP("since", "s", 3*time.Hour, "from what time to begin displaying logs, default is 3 hours ago")
	cmd.Flags().BoolP("tail", "f", false, "continuously poll for new logs")
	cmd.Flags().String("stage", config.DefaultStageName, "stage name")
	rootCmd.AddCommand(cmd)
}

func addCommandNew() {
	cmd := &cobra.Command{
		Use:   "new <project>",
		Short: "Initializes a new Mantil project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			c := initNew(cmd, args)
			if err := c.run(); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().String("from", "", "name of the template or URL of the repository that will be used as one")
	cmd.Flags().String("module-name", "", "replace module name and import paths")
	rootCmd.AddCommand(cmd)
}

func addCommandTest() {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run project integration tests",
		Long: `Run project integration tests

Project integration tests are pure Go test in [project-root]/test folder.
Mantil sets MANTIL_API_URL environment variable to point to the current
project api url and runs tests with 'go test -v'.
`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			c := initTest(cmd, args)
			if err := c.run(); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().StringP("run", "r", "", "run only tests with this pattern in name")
	cmd.Flags().StringP("stage", "s", config.DefaultStageName, "stage name")
	rootCmd.AddCommand(cmd)

}

func addCommandWatch() {
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch for file changes and automatically deploy functions",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			c := initWatch(cmd, args)
			if err := c.run(); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().BoolP("test", "t", false, "run tests after deploying changes")
	cmd.Flags().StringP("method", "m", "", "method to invoke after deploying changes")
	cmd.Flags().StringP("data", "d", "", "data for the method invoke request")
	cmd.Flags().StringP("stage", "s", config.DefaultStageName, "stage name")
	rootCmd.AddCommand(cmd)
}

func initDestroy(cmd *cobra.Command, args []string) *destroyCmd {
	force, _ := cmd.Flags().GetBool("force")
	stageName, _ := cmd.Flags().GetString("stage")
	deleteRepo, _ := cmd.Flags().GetBool("repo")

	ctx := commands.MustProjectContextWithStage(stageName)
	if !force {
		confirmProjectDestroy(ctx.Project)
	}

	return &destroyCmd{
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

	ctx := commands.MustProjectContextWithStage(stageName)

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

	ctx := commands.MustProjectContextWithStage(stageName)
	awsClient := ctx.InitialiseAWSSDK()

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

func initWatch(cmd *cobra.Command, args []string) *watchCmd {
	method := cmd.Flag("method").Value.String()
	test, _ := cmd.Flags().GetBool("test")
	data := cmd.Flag("data").Value.String()
	stageName, _ := cmd.Flags().GetString("stage")

	ctx := commands.MustProjectContextWithStage(stageName)
	awsClient := ctx.InitialiseAWSSDK()

	deploy, err := deploy.New(ctx, awsClient)
	if err != nil {
		log.Fatal(err)
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

func confirmProjectDestroy(p *config.Project) {
	confirmationPrompt := promptui.Prompt{
		Label: "To confirm deletion, please enter the project name",
	}
	projectName, err := confirmationPrompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	if p.Name != projectName {
		log.Fatalf("Project name doesn't match, exiting...")
	}
}

func selectFunctionFromStage(stage *config.Stage) string {
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
		log.Fatal(err)
	}
	return function
}
