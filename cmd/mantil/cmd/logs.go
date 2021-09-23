package cmd

import (
	"log"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/internal/cli/commands/logs"
	"github.com/mantil-io/mantil/internal/mantil"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs [function]",
	Short: "Fetch logs for a specific function/api",
	Long: `Fetch logs for a specific function/api

For the description of filter patterns see:
https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		since, _ := cmd.Flags().GetDuration("since")
		filter := cmd.Flag("filter-pattern").Value.String()
		tail, _ := cmd.Flags().GetBool("follow")

		config, _, token := localData()
		p := fetchProject(config.Name, token)

		var function string
		if len(args) > 0 {
			function = args[0]
		} else {
			function = selectFunction(p)
		}

		logGroup := mantil.ProjectResource(p.Name, function)
		aws := initialiseAWSSDK(config.Name, token)
		l := logs.New(aws)
		if err := l.Fetch(logGroup, filter, since, tail); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	logsCmd.Flags().StringP("filter-pattern", "p", "", "filter pattern to use")
	logsCmd.Flags().DurationP("since", "s", 3*time.Hour, "from what time to begin displaying logs, default is 3 hours ago")
	logsCmd.Flags().BoolP("follow", "f", false, "continuously poll for new logs")
	rootCmd.AddCommand(logsCmd)
}

func selectFunction(p *mantil.Project) string {
	var funcNames []string
	for _, f := range p.Functions {
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
