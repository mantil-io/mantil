package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/commands/invoke"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/spf13/cobra"
)

var invokeCmd = &cobra.Command{
	Use:   "invoke <function>[/method]",
	Short: "Makes requests to functions through project's API Gateway",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p, _ := getProject()
		stageName, err := cmd.Flags().GetString("stage")
		if err != nil {
			log.Fatal(err)
		}
		if p.RestEndpoint(stageName) == "" {
			log.Fatalf("api URL for the stage does not exist")
		}
		data := cmd.Flag("data").Value.String()
		includeHeaders, err := cmd.Flags().GetBool("include")
		if err != nil {
			includeHeaders = false
		}
		includeLogs, err := cmd.Flags().GetBool("logs")
		if err != nil {
			includeLogs = false
		}
		endpoint := fmt.Sprintf("%s/%s", p.RestEndpoint(stageName), args[0])
		if err := invoke.Endpoint(endpoint, data, includeHeaders, includeLogs); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	invokeCmd.Flags().StringP("data", "d", "", "data for the method invoke request")
	invokeCmd.Flags().BoolP("include", "i", false, "include response headers in the output")
	invokeCmd.Flags().BoolP("logs", "l", false, "show lambda execution logs")
	invokeCmd.Flags().StringP("stage", "s", config.DefaultStageName, "stage name")
	rootCmd.AddCommand(invokeCmd)
}
