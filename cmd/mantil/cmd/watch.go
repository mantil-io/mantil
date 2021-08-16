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
		d, err := deploy.New(p, aws, path, token)
		if err != nil {
			log.Fatal(err)
		}
		watch.Start(path, func() {
			log.Println("changes detected - starting deploy")
			defer log.Println("deploy successfully finished")
			if err := d.Deploy(); err != nil {
				log.Fatal(err)
			}
		})
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
