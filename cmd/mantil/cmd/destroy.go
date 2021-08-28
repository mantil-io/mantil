package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/internal/cli/commands/destroy"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy all infrastructure resources",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, path, token := localData()
		p := fetchProject(config.Name, token)

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
		deleteRepo, err := cmd.Flags().GetBool("repo")
		if err != nil {
			log.Fatal(err)
		}
		if err := d.Destroy(deleteRepo); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	destroyCmd.Flags().Bool("repo", false, "delete Github repo and local code folder")
	rootCmd.AddCommand(destroyCmd)
}
