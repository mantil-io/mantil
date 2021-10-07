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
		Short: "Deploys updates to stages",
		Long: `Deploys updates to stages

This command checks if any assets, code or configuration have changed since the last deployment
and applies the necessary updates.

The --stage flag accepts any existing stage and defaults to the default stage if omitted.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stageName, err := cmd.Flags().GetString("stage")
			if err != nil {
				return log.Wrap(err)
			}
			ctx, err := project.NewContext()
			if err != nil {
				return log.Wrap(err)
			}
			stage := ctx.ResolveStage(stageName)
			if stage == nil {
				return log.WithUserMessage(nil, "The specified stage doesn't exist, create it with `mantil stage new`.")
			}
			ctx.SetStage(stage)
			awsClient, err := ctx.AWSClient()
			if err != nil {
				return log.Wrap(err)
			}
			d, err := deploy.New(ctx, awsClient)
			if err != nil {
				return log.Wrap(err)
			}
			if _, err = d.Deploy(); err != nil {
				return log.Wrap(err)
			}
			return nil
		},
	}
	cmd.Flags().StringP("stage", "s", "", "the name of the stage to deploy to")
	return cmd
}
