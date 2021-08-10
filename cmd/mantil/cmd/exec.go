package cmd

import (
	"fmt"
	"log"

	"github.com/atoz-technology/mantil-cli/internal/commands/exec"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Makes request to functions through project's gateway api.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p, _, _, _ := findProject([]string{})
		data := cmd.Flag("data").Value.String()
		endpoint := fmt.Sprintf("%s/%s", p.ApiURL, args[0])

		if err := exec.Endpoint(endpoint, data); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	execCmd.Flags().StringP("data", "d", "", "Data for the request")
	rootCmd.AddCommand(execCmd)
}
