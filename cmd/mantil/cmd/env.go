package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil-cli/internal/mantil"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Show project environment variables",
	Long: `Show project environment variables

You can set environment variables in terminal with:
$ eval $(mantil env)
`,
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetBool("url")
		env, config := mantil.Env()
		if url {
			fmt.Printf("%s", config.ApiURL)
			return
		}
		fmt.Printf("%s", env)

	},
}

func init() {
	envCmd.Flags().BoolP("url", "u", false, "show only project api url")
	rootCmd.AddCommand(envCmd)
}
