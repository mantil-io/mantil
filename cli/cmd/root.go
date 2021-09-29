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
	rootCmd.AddCommand(newDestroyCommand())
	rootCmd.AddCommand(newEnvCommand())
	rootCmd.AddCommand(newInvokeCommand())
	rootCmd.AddCommand(newLogsCommand())
	rootCmd.AddCommand(newNewCommand())
	rootCmd.AddCommand(newTestCommand())
	rootCmd.AddCommand(newWatchCommand())
	rootCmd.AddCommand(newSetupCommand())
	rootCmd.AddCommand(newDeployCommand())
	rootCmd.AddCommand(newGenerateCommand())

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
