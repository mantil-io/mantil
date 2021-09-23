package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/mantil"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show project environment variables",
	Long: `Show project environment variables

You can set environment variables in terminal with:
$ eval $(mantil env)
`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetBool("url")
		stageName, _ := cmd.Flags().GetString("stage")
		env, stage := mantil.Env(stageName)
		if url && stage != nil && stage.Endpoints != nil {
			fmt.Printf("%s", stage.Endpoints.Rest)
			return
		}
		fmt.Printf("%s", env)

	},
}

func init() {
	envCmd.Flags().BoolP("url", "u", false, "show only project api url")
	envCmd.Flags().StringP("stage", "s", mantil.DefaultStageName, "stage name")
	rootCmd.AddCommand(envCmd)
}
