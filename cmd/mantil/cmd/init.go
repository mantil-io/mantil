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
		namePrompt := promptui.Prompt{
			Label: "Project name",
		}
		projectName, err := namePrompt.Run()
		if err != nil {
			log.Fatal(err)
		}
		i, err := initialize.New(projectName)
		if err != nil {
			log.Fatal(err)
		}
		if err := i.InitProject(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
