package cmd

import (
	"fmt"

	"github.com/atoz-technology/mantil-cli/internal/commands/deploy"
	"github.com/atoz-technology/mantil-cli/internal/commands/invoke"
	"github.com/atoz-technology/mantil-cli/internal/commands/watch"
	"github.com/atoz-technology/mantil-cli/internal/log"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for file changes and automatically deploy functions.",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		p, config, path, token := findProject(args)
		method := cmd.Flag("method").Value.String()
		data := cmd.Flag("data").Value.String()

		if method != "" && p.ApiURL == "" {
			log.Fatalf("api URL for the project does not exist")
		}
		endpoint := fmt.Sprintf("%s/%s", p.ApiURL, method)

		aws, err := initialiseAWSSDK(config.Name, token)
		if err != nil {
			log.Fatal(err)
		}

		d, err := deploy.New(p, aws, path, token)
		if err != nil {
			log.Fatal(err)
		}

		watch.Start(path, func() {
			log.Info("changes detected - starting deploy")
			if err := d.Deploy(); err != nil {
				log.Fatal(err)
			}
			if method != "" {
				log.Info("invoking method %s", method)
				if err := invoke.Endpoint(endpoint, data, false, true); err != nil {
					log.Error(err)
				}
			}
		})
	},
}

func init() {
	watchCmd.Flags().StringP("method", "m", "", "Method to invoke after deploying changes.")
	watchCmd.Flags().StringP("data", "d", "", "Data for the method request.")
	rootCmd.AddCommand(watchCmd)
}
