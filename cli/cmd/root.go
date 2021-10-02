package cmd

import (
	"context"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/spf13/cobra"
)

func Execute(ctx context.Context, version string) error {
	var cmd = &cobra.Command{
		Use:           "mantil",
		Short:         "Makes serverless development with Go and AWS Lambda joyful",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// no-color is handled in cli/log pacakge
	cmd.PersistentFlags().Bool("no-color", false, "don't use colors in output")

	add := func(factory func() *cobra.Command) {
		sub := factory()
		cmd.AddCommand(sub)
	}
	subCommands := []func() *cobra.Command{
		newDestroyCommand,
		newEnvCommand,
		newInvokeCommand,
		newLogsCommand,
		newNewCommand,
		newTestCommand,
		newWatchCommand,
		newDeployCommand,
		newGenerateCommand,
		newAwsCommand,
		// for testing:
		//newErrorsExample,
	}
	for _, sub := range subCommands {
		add(sub)
	}

	ec, err := cmd.ExecuteContextC(ctx)
	if err == nil {
		return nil
	}

	ui.Error(err)              // this will handle UserError case
	if !log.IsUserError(err) { // show usage for cobar errors, and other usage errors
		ec.Usage()
	}
	return err
}
