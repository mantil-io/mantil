package cmd

import (
	"github.com/mantil-io/mantil/cli/mantil/commands/initialize"
	"github.com/mantil-io/mantil/cli/mantil/log"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <project>",
	Short: "Initializes a new Mantil project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		repo := cmd.Flag("from").Value.String()
		moduleName := cmd.Flag("module-name").Value.String()
		if moduleName == "" {
			moduleName = projectName
		}

		i, err := initialize.New(projectName, repo, moduleName)
		if err != nil {
			log.Fatal(err)
		}
		if err := i.InitProject(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	newCmd.Flags().String("from", "", "name of the template or URL of the repository that will be used as one")
	newCmd.Flags().String("module-name", "", "replace module name and import paths")
	rootCmd.AddCommand(newCmd)
}
