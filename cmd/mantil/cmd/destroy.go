package cmd

import (
	"log"

	"github.com/atoz-technology/mantil-cli/internal/commands/destroy"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:  "destroy",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		p, org, path := findProject(args)
		d, err := destroy.New(p, org, path)
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
			log.Fatal("Project name doesn't match, exiting...")
		}
		if err := d.Destroy(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
