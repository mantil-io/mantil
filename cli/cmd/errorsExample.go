package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/spf13/cobra"
)

func newErrorsExample() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "try",
		Short: "try some development concept",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pero := cmd.Flag("pero").Value.String()
			var es errStack
			return es.run(pero)
		},
	}
	cmd.Flags().StringP("pero", "", "", "pero ne moze biti zdero")
	return cmd
}

type errStack struct{}

func (e *errStack) run(pero string) error {
	if pero == "zdero" { // this will show error and command usage!
		return log.Wrap(fmt.Errorf("pero ne moze biti zdero"))
	}
	ui.Info("in run")
	return e.first()
}

func (e *errStack) first() error {
	ui.Notice("in first")
	if err := e.second(); err != nil {
		return log.Wrap(err, "first got error")
	}
	return nil
}

func (e *errStack) second() error {
	ui.Debug("in second")
	if err := e.third(); err != nil {
		return log.WithUserMessage(err, "message that should be shown to the user")
	}
	return nil
}

func (e *errStack) third() error {
	ui.Errorf("in third")
	return log.Wrap(fmt.Errorf("third failed"))
}
