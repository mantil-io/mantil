package cmd

import (
	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/spf13/cobra"
)

func newDeployCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Creates infrastructure and deploys updates to lambda functions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stageName, err := cmd.Flags().GetString("stage")
			if err != nil {
				return err
			}
			ctx := project.MustContext()
			stage := ctx.ResolveStage(stageName)
			if stage == nil {
				return log.WithUserMessage(nil, "The specified stage doesn't exist, create it with `mantil stage new`.")
			}
			ctx.SetStage(stage)
			aws := ctx.MustInitialiseAWSSDK()

			d, err := deploy.New(ctx, aws)
			if err != nil {
				return err
			}
			if _, err = d.Deploy(); err != nil {
				return err

			}
			return nil
		},
	}
	cmd.Flags().StringP("stage", "s", "", "name of the stage to deploy to")
	return cmd
}
