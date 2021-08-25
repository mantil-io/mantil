package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil-cli/internal/commands/logs"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Fetch logs for a specific function/api",
	Long: `Fetch logs for a specific function/api

For the description of filter patterns see:
https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html`,
	Run: func(cmd *cobra.Command, args []string) {
		function := cmd.Flag("name").Value.String()
		since, _ := cmd.Flags().GetDuration("since")
		filter := cmd.Flag("filter-pattern").Value.String()
		tail, _ := cmd.Flags().GetBool("follow")
		p, config, _, token := findProject(args)
		if function == "" {
			var funcNames []string
			for _, f := range p.Functions {
				funcNames = append(funcNames, f.Name)
			}
			prompt := promptui.Select{
				Label: "Select a function",
				Items: funcNames,
			}
			var err error
			_, function, err = prompt.Run()
			if err != nil {
				log.Fatal(err)
			}
		}
		function = fmt.Sprintf("%s-%s", p.Name, function)
		aws, err := initialiseAWSSDK(config.Name, token)
		if err != nil {
			log.Fatal(err)
		}
		l := logs.New(aws)
		if err := l.Fetch(function, filter, since, tail); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	logsCmd.Flags().StringP("name", "n", "", "function/api name to fetch logs for")
	logsCmd.Flags().StringP("filter-pattern", "p", "", "filter pattern to use")
	logsCmd.Flags().DurationP("since", "s", 3*time.Hour, "from what time to begin displaying logs, default is 3 hours ago")
	logsCmd.Flags().BoolP("follow", "f", false, "continuously poll for new logs")
	rootCmd.AddCommand(logsCmd)
}
