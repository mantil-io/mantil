package cmd

import (
	"github.com/mantil-io/mantil/internal/cli/commands/deploy"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Creates infrastructure and deploys updates to lambda functions",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		config, path, token := localData()
		p := fetchProject(config.Name, token)
		aws := initialiseAWSSDK(config.Name, token)

		d, err := deploy.New(p, config, aws, path, token)
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
