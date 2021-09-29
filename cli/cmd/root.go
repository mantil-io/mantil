package cmd

import (
	"github.com/mantil-io/mantil/cli/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mantil",
	Short: "Makes serverless development with Go and AWS Lambda joyful",
}

func Root() *cobra.Command {
	return rootCmd
}

func init() {
	var verbose, noColor bool
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose log output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "don't use color in log output")

	cobra.OnInitialize(func() {
		if verbose {
			log.EnableDebugLogLevel()
		}
		if noColor {
			log.DisableColor()
		}
	})
}
