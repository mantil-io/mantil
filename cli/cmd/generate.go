package cmd

import (
	"github.com/mantil-io/mantil/cli/cmd/generate"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/spf13/cobra"
)

func newGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Automatically generate code in the project",
	}
	cmd.AddCommand(newGenerateApiCommand())
	return cmd
}

func newGenerateApiCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api <function>",
		Short: "Generate Go code for new api",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			methods, err := cmd.Flags().GetStringSlice("methods")
			if err != nil {
				return err
			}
			if err := generate.Api(name, methods); err != nil {
				return err
			}
			log.Notice("successfuly generated api %s", name)
			return nil
		},
	}
	cmd.Flags().StringSliceP("methods", "m", nil, "additional function methods, if left empty only the Default method will be created")
	return cmd
}
