package cmd

import (
	"context"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/spf13/cobra"
)

func Execute(ctx context.Context, version string) error {
	var cmd = &cobra.Command{
		Use:     "mantil",
		Short:   "Makes serverless development with Go and AWS Lambda joyful",
		Version: version,
	}
	// add sub commands
	cmd.AddCommand(newDestroyCommand())
	cmd.AddCommand(newEnvCommand())
	cmd.AddCommand(newInvokeCommand())
	cmd.AddCommand(newLogsCommand())
	cmd.AddCommand(newNewCommand())
	cmd.AddCommand(newTestCommand())
	cmd.AddCommand(newWatchCommand())
	cmd.AddCommand(newSetupCommand())
	cmd.AddCommand(newDeployCommand())
	cmd.AddCommand(newGenerateCommand())

	// register global flags
	var verbose, noColor bool
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose log output")
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "don't use color in log output")
	// use global flags
	cobra.OnInitialize(func() {
		if verbose {
			log.EnableDebugLogLevel()
		}
		if noColor {
			log.DisableColor()
		}
	})

	return cmd.ExecuteContext(ctx)
}
