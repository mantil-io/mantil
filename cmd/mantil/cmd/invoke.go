package cmd

import (
	"fmt"
	"log"

	"github.com/atoz-technology/mantil-cli/internal/commands/invoke"
	"github.com/spf13/cobra"
)

var invokeCmd = &cobra.Command{
	Use:   "invoke",
	Short: "Makes requests to functions through project's gateway api.",
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
		endpoint := fmt.Sprintf("%s/%s", p.ApiURL, args[0])

		if err := invoke.Endpoint(endpoint, data, includeHeaders); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	invokeCmd.Flags().StringP("data", "d", "", "Data for the request")
	invokeCmd.Flags().BoolP("include", "i", false, "Include response headers in the output.")
	rootCmd.AddCommand(invokeCmd)
}
