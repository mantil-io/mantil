package cmd

import (
	"github.com/mantil-io/mantil-cli/internal/commands/destroy"
	"github.com/mantil-io/mantil-cli/internal/log"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:  "destroy",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p, config, path, token := findProject(args)
		d, err := destroy.New(p, config.GithubOrg, path, token)
		if err != nil {
			log.Fatal(err)
		}
		confirmationPrompt := promptui.Prompt{
			Label: "To confirm deletion, please enter the project name",
		}
		projectName, err := confirmationPrompt.Run()
		if err != nil {
			log.Fatal(err)
		}
		if p.Name != projectName {
			log.Fatalf("Project name doesn't match, exiting...")
		}
		if err := d.Destroy(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
