package cmd

import (
	"log"

	"github.com/atoz-technology/mantil-cli/internal/commands/deploy"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Creates infrastructure and deploys updates to lambda functions",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		d, err := deploy.New(findProject(args))
		if err != nil {
			log.Fatal(err)
		}
		if err = d.Deploy(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
