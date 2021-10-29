package examples

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/controller"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/event"
	"github.com/spf13/cobra"
)

func NewErrorsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "try",
		Short: "try some development concept",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pero := cmd.Flag("pero").Value.String()
			var es errStack
			err := es.run(pero)
			log.Event(event.Event{Deploy: &event.Deploy{BuildDuration: 1}})
			if err != nil {
				return log.Wrap(err, "high level message")
			}
			return nil
		},
	}
	cmd.Flags().StringP("pero", "", "", "pero ne moze biti zdero")
	return cmd
}

type errStack struct{}

func (e *errStack) run(pero string) error {
	if pero == "zdero" { // this will show error and command usage!
		//return log.Wrap(fmt.Errorf("pero ne moze biti zdero"))
		return log.Wrap(controller.NewArgumentError("pero ne moze biti zdero"))
	}
	return e.first()
}

func (e *errStack) first() error {
	if err := e.second(); err != nil {
		return log.Wrap(err, "first got error")
	}
	return nil
}

func (e *errStack) second() error {
	if err := e.third(); err != nil {
		return log.Wrap(err, "message that should be shown to the user")
	}
	return nil
}

func (e *errStack) third() error {
	return log.Wrap(fmt.Errorf("third failed"))
}
