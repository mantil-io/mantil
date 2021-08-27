package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil-cli/internal/commands/initialize"
	"github.com/mantil-io/mantil-cli/internal/log"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new Mantil project",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		projectName, err := promptProjectName()
		if err != nil {
			log.Fatalf("could not prompt project name - %v", err)
		}
		noRepo, _ := cmd.Flags().GetBool("no-repo")
		template := cmd.Flag("template").Value.String()
		var githubOrg string
		if !noRepo {
			githubOrg, err = promptGithubOrganization()
			if err != nil {
				log.Fatalf("could not prompt github organization - %v", err)
			}
		}
		i, err := initialize.New(projectName, githubOrg, noRepo, template)
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
	initCmd.Flags().Bool("no-repo", false, "Skip creating a github repository for the project")
	initCmd.Flags().StringP("template", "t", "", "project template name, one of: ping, excuses")
	rootCmd.AddCommand(initCmd)
}
