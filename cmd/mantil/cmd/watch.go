package cmd

import (
	"log"

	"github.com/atoz-technology/mantil-cli/internal/commands/deploy"
	"github.com/atoz-technology/mantil-cli/internal/commands/watch"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for file changes and automatically deploy functions.",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		p, config, path, token := findProject(args)
		aws, err := initialiseAWSSDK(config.Name, token)
		if err != nil {
			log.Fatal(err)
		}
		whitelistedFuncs, err := cmd.Flags().GetStringSlice("functions")
		if err != nil {
			log.Fatal(err)
		}
		d, err := deploy.New(p, aws, path, token, whitelistedFuncs...)
		if err != nil {
			log.Fatal(err)
		}
		watch.Start(path, func() {
			if err := d.Deploy(); err != nil {
				log.Fatal(err)
			}
		})
	},
}

func init() {
	watchCmd.Flags().StringSliceP("functions", "f", nil, "Whitelist functions which will be deployed on changes. If left empty all are whitelisted.")
	rootCmd.AddCommand(watchCmd)

}
