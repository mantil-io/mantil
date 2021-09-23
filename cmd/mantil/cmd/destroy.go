package cmd

import (
	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/internal/cli/commands/destroy"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/mantil"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy all infrastructure resources",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		config, path, token := localData()
		p := fetchProject(config.Name, token)
		confirmProjectDestroy(p)
		d, err := destroy.New(p, path, token)
		if err != nil {
			log.Fatal(err)
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

func confirmProjectDestroy(p *mantil.Project) {
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
}

func init() {
	destroyCmd.Flags().Bool("repo", false, "delete local repository")
	rootCmd.AddCommand(destroyCmd)
}
