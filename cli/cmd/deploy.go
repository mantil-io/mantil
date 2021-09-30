package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/workspace"
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
				ctx.SetStage(createStage(stageName, ctx))
			} else {
				ctx.SetStage(stage)
			}
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
	cmd.Flags().StringP("stage", "s", "", "name of the stage to deploy to, if the stage doesn't exist yet it will be created")
	return cmd
}

func createStage(stageName string, ctx *project.Context) (stage *workspace.Stage) {
	if len(ctx.Workspace.Accounts) == 0 {
		log.Fatalf("No accounts found in workspace. Please set up an account with mantil setup.")
	}
	if stageName == "" {
		stageName = workspace.DefaultStageName
	}
	var accountName string
	if len(ctx.Workspace.Accounts) > 1 {
		accountName = selectAccount(ctx.Workspace)
	} else {
		accountName = ctx.Workspace.Accounts[0].Name
	}
	stage = &workspace.Stage{
		Name:    stageName,
		Account: accountName,
	}
	if len(ctx.Project.Stages) == 0 {
		stage.Default = true
	}
	return stage
}

func selectAccount(w *workspace.Workspace) string {
	var accounts []string
	for _, a := range w.Accounts {
		accounts = append(accounts, a.Name)
	}
	prompt := promptui.Select{
		Label: "Select an account",
		Items: accounts,
	}
	_, account, err := prompt.Run()
	if err != nil {
		log.Fatal(err)
	}
	return account
}
