package cmd

import (
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/spf13/cobra"
)

func init() {
	addCommandEnv()
	addCommandInvoke()
	addCommandLogs()
	addCommandNew()
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

func initEnv(cmd *cobra.Command, args []string) *envCmd {
	url, _ := cmd.Flags().GetBool("url")
	stageName, _ := cmd.Flags().GetString("stage")

	return &envCmd{
		url:       url,
		stageName: stageName,
	}
}

func initInvoke(cmd *cobra.Command, args []string) *invokeCmd {
	p, _ := getProject()
	stageName, _ := cmd.Flags().GetString("stage")
	data := cmd.Flag("data").Value.String()
	includeHeaders, _ := cmd.Flags().GetBool("include")
	includeLogs, _ := cmd.Flags().GetBool("logs")

	return &invokeCmd{
		endpoint:       args[0],
		project:        p,
		stageName:      stageName,
		data:           data,
		includeHeaders: includeHeaders,
		includeLogs:    includeLogs,
	}
}

func initLogs(cmd *cobra.Command, args []string) *logsCmd {
	p, _ := getProject()
	stageName, _ := cmd.Flags().GetString("stage")
	awsClient := initialiseAWSSDK(p.Name, stageName)
	filter := cmd.Flag("filter-pattern").Value.String()
	since, _ := cmd.Flags().GetDuration("since")
	tail, _ := cmd.Flags().GetBool("tail")

	stage := p.Stage(stageName)
	if stage == nil {
		log.Fatalf("stage %s not found", stageName)
	}

	var function string
	if len(args) > 0 {
		function = args[0]
	} else {
		function = selectFunctionFromStage(stage)
	}
	startTime := time.Now().Add(-since)

	return &logsCmd{
		project:   p,
		stageName: stageName,
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
