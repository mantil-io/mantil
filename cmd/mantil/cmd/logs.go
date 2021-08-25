package cmd

import (
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil-cli/internal/commands/logs"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Fetch logs for a specific function/api",
	Run: func(cmd *cobra.Command, args []string) {
		function := cmd.Flag("function").Value.String()
		start := cmd.Flag("start").Value.String()
		filter := cmd.Flag("filter").Value.String()
		tail, _ := cmd.Flags().GetBool("tail")
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
		if err := l.Fetch(function, filter, start, tail); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	logsCmd.Flags().StringP("name", "n", "", "function/api name to fetch logs for")
	logsCmd.Flags().String("filter-pattern", "p", "filter pattern to use see (https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/FilterAndPatternSyntax.html)")
	logsCmd.Flags().StringP("since", "s", "", "from what time to begin displaying logs, default is 3 hours ago")
	logsCmd.Flags().BoolP("follow", "f", false, "continuously poll for new logs")
	rootCmd.AddCommand(logsCmd)
}
