package cmd

import (
	"log"
	"os"

	"github.com/atoz-technology/mantil-cli/internal/commands/deploy"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Creates infrastructure and deploys updates to lambda functions",
	Run: func(cmd *cobra.Command, args []string) {
		projectRoot, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		d, err := deploy.New(projectRoot)
		if err != nil {
			log.Fatal(err)
		}
		if err := d.UploadFunctions(); err != nil {
			log.Fatal(err)
		}
		if err := d.CreateInfrastructure(); err != nil {
			log.Fatal(err)
		}
		if err := d.UpdateLambdaFunctions(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
