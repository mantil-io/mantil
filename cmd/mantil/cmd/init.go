package cmd

import (
	"log"

	"github.com/atoz-technology/mantil-cli/internal/commands/initialize"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a mantil project",
	Run: func(cmd *cobra.Command, args []string) {
		projectName, err := promptProjectName()
		if err != nil {
			log.Fatalf("could not prompt project name - %v", err)
		}
		githubOrg, err := promptGithubOrganization()
		if err != nil {
			log.Fatalf("could not prompt github organization - %v", err)
		}
		i, err := initialize.New(projectName, githubOrg)
		if err != nil {
			log.Fatal(err)
		}
		if err := i.InitProject(); err != nil {
			log.Fatal(err)
		}
	},
}

func promptProjectName() (string, error) {
	namePrompt := promptui.Prompt{
		Label: "Project name",
	}
	return namePrompt.Run()
}

func promptGithubOrganization() (string, error) {
	orgPrompt := promptui.Prompt{
		Label: "Github organization",
	}
	return orgPrompt.Run()
}

func init() {
	rootCmd.AddCommand(initCmd)
}
