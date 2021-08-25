package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil-cli/internal/log"

	"github.com/mantil-io/mantil-cli/internal/commands/invoke"
	"github.com/spf13/cobra"
)

var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "Makes requests to functions through project's API Gateway",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p, _, _, _ := findProject([]string{})
		if p.ApiURL == "" {
			log.Fatalf("api URL for the project does not exist")
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
		endpoint := fmt.Sprintf("%s/%s", p.ApiURL, args[0])
		if err := invoke.Endpoint(endpoint, data, includeHeaders, includeLogs); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	invokeCmd.Flags().StringP("data", "d", "", "data for the method invoke request")
	invokeCmd.Flags().BoolP("include", "i", false, "include response headers in the output")
	invokeCmd.Flags().BoolP("logs", "l", false, "show lambda execution logs")
	rootCmd.AddCommand(invokeCmd)
}
