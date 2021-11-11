package examples

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/controller"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/domain"
	"github.com/spf13/cobra"
)

func NewErrorsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "errors",
		Short: "try some development concept",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pero := cmd.Flag("pero").Value.String()
			var es errStack
			err := es.run(pero)
			log.Event(domain.Event{Deploy: &domain.Deploy{BuildDuration: 1}})
			if err != nil {
				return log.Wrap(err, "high level message")
				//return log.Wrapf("only one message")
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
		return log.Wrap(err, "first wrapped with message")
	}
	return nil
}

func (e *errStack) second() error {
	if err := e.third(); err != nil {
		return log.Wrap(err, "second wrapped with message")
	}
	return nil
}

func (e *errStack) third() error {
	return log.Wrap(fmt.Errorf("third failed"))
	//return log.Wrapf("third failed")
}
