package cmd

import (
	"github.com/mantil-io/mantil/cli/build"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func Execute() error {
	ec, err := root().ExecuteC()
	if err == nil {
		return nil
	}

	ui.Error(err)              // this will handle UserError case
	if !log.IsUserError(err) { // show usage for cobar errors, and other usage errors
		ec.Usage()
	}
	return err
}

func root() *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "mantil",
		Short:         "Makes serverless development with Go and AWS Lambda joyful",
		Version:       build.Version().String(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// no-color is handled in cli/log pacakge
	cmd.PersistentFlags().Bool("no-color", false, "don't use colors in output")
	cmd.PersistentFlags().Bool("help", false, "show command help") // move help to global commands
	cmd.Flags().Bool("version", false, "show mantil version")      // remove -v shortcut for version

	add := func(factory func() *cobra.Command) {
		sub := factory()
		cmd.AddCommand(sub)
	}
	subCommands := []func() *cobra.Command{
		newEnvCommand,
		newInvokeCommand,
		newLogsCommand,
		newNewCommand,
		newTestCommand,
		newWatchCommand,
		newDeployCommand,
		newGenerateCommand,
		newAwsCommand,
		newStageCommand,

		// for testing:
		//examples.NewErrorsCommand,
		//examples.NewArgsCommand,
	}
	for _, sub := range subCommands {
		add(sub)
	}
	return cmd
}

func GenDoc(dir string) error {
	cmd := root()
	cmd.DisableAutoGenTag = true
	return doc.GenMarkdownTree(cmd, dir)
}
