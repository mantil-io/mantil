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
		p, path := getProject()
		stage, isNew := resolveStage(cmd, p)
		stageName := stage.Name
		if isNew {
			stageName = ""
		}
		aws := initialiseAWSSDK(p.Name, stageName)
		account := getAccount(stage.Name)
		d, err := deploy.New(account, p, stage, aws, path)
		if err != nil {
			log.Fatal(err)
		}
		if _, err = d.Deploy(); err != nil {
			log.Fatal(err)
		}
	},
}

func resolveStage(cmd *cobra.Command, p *config.Project) (stage *config.Stage, isNew bool) {
	w, err := commands.LoadWorkspaceConfig()
	if err != nil {
		log.Fatal(err)
	}
	if len(w.Accounts) == 0 {
		log.Fatalf("No accounts found in workspace. Please set up an account with mantil setup.")
	}
	stageName, err := cmd.Flags().GetString("stage")
	if err != nil {
		log.Fatal(err)
	}
	if stageName == "" {
		stageName = config.DefaultStageName
	}
	if s := p.Stage(stageName); s != nil {
		return s, false
	}
	// if the stage doesn't exist create it
	var accountName string
	if len(w.Accounts) > 1 {
		accountName = selectAccount(w)
	} else {
		accountName = w.Accounts[0].Name
	}
	stage = &config.Stage{
		Name:    stageName,
		Account: accountName,
	}
	return stage, true
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
	deployCmd.Flags().StringP("stage", "s", "", "stage name")
}
