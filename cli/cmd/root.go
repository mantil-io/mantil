package cmd

import (
	"context"

	"github.com/mantil-io/mantil/cli/log"
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

		//newErrorsExample,
	}
	for _, sub := range subCommands {
		add(sub)
	}

	// register global flags
	var noColor bool
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "don't use color in log output")
	// use global flags
	cobra.OnInitialize(func() {
		if noColor {
			log.UI.DisableColor()
		}
	})

	ec, err := cmd.ExecuteContextC(ctx)
	if err == nil {
		return nil
	}

	log.UI.Error(err)          // this will handle UserError case
	if !log.IsUserError(err) { // show usage for cobar errors, and other usage errors
		ec.Usage()
	}
	return err // signal main to set non-zero exit code
}
