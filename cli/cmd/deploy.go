package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/commands"
	"github.com/mantil-io/mantil/cli/commands/deploy"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Creates infrastructure and deploys updates to lambda functions",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		stageName, err := cmd.Flags().GetString("stage")
		if err != nil {
			log.Fatal(err)
		}
		ctx, stageExists := commands.MustProjectContext(stageName)
		if !stageExists {
			ctx.SetStage(createStage(stageName, ctx))
		}
		aws := ctx.InitialiseAWSSDK()

		d, err := deploy.New(ctx, aws)
		if err != nil {
			log.Fatal(err)
		}
		if _, err = d.Deploy(); err != nil {
			log.Fatal(err)
		}
	},
}

func createStage(stageName string, ctx *commands.ProjectContext) (stage *config.Stage) {
	if len(ctx.Workspace.Accounts) == 0 {
		log.Fatalf("No accounts found in workspace. Please set up an account with mantil setup.")
	}
	if stageName == "" {
		stageName = config.DefaultStageName
	}
	var accountName string
	if len(ctx.Workspace.Accounts) > 1 {
		accountName = selectAccount(ctx.Workspace)
	} else {
		accountName = ctx.Workspace.Accounts[0].Name
	}
	stage = &config.Stage{
		Name:    stageName,
		Account: accountName,
	}
	return stage
}

func selectAccount(w *commands.WorkspaceConfig) string {
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

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringP("stage", "s", config.DefaultStageName, "stage name")
}
