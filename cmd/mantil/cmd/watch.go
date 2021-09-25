package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/cmd/mantil/commands/deploy"
	"github.com/mantil-io/mantil/cmd/mantil/commands/invoke"
	"github.com/mantil-io/mantil/cmd/mantil/commands/watch"
	"github.com/mantil-io/mantil/cmd/mantil/log"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/shell"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for file changes and automatically deploy functions",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		p, path := getProject()

		method := cmd.Flag("method").Value.String()
		test, _ := cmd.Flags().GetBool("test")
		data := cmd.Flag("data").Value.String()
		stageName, _ := cmd.Flags().GetString("stage")

		stage := p.Stage(stageName)
		if stage == nil {
			log.Fatalf("invalid stage name")
		}
		if method != "" && p.RestEndpoint(stageName) == "" {
			log.Fatalf("api URL for the stage does not exist")
		}
		endpoint := fmt.Sprintf("%s/%s", p.RestEndpoint(stageName), method)
		aws := initialiseAWSSDK(p.Name, stage.Name)

		d, err := deploy.New(p, stage, aws, path)
		if err != nil {
			log.Fatal(err)
		}

		watch.Start(path, func() {
			log.Info("\nchanges detected - starting deploy")
			updated, err := d.Deploy()
			if err != nil {
				log.Fatal(err)
			}
			if !updated {
				return
			}
			if method != "" {
				log.Info("invoking method %s", method)
				if err := invoke.Endpoint(endpoint, data, false, true); err != nil {
					log.Error(err)
				}
			}
			if test {
				log.Info("running tests")
				err := shell.Exec(shell.ExecOptions{
					Args:    []string{"go", "test", "-v"},
					WorkDir: path + "/test",
					Logger:  log.Info,
				})
				if err != nil {
					log.Error(err)
				}
			}
		})
	},
}

func init() {
	watchCmd.Flags().BoolP("test", "t", false, "run tests after deploying changes")
	watchCmd.Flags().StringP("method", "m", "", "method to invoke after deploying changes")
	watchCmd.Flags().StringP("data", "d", "", "data for the method invoke request")
	watchCmd.Flags().StringP("stage", "s", config.DefaultStageName, "stage name")
	rootCmd.AddCommand(watchCmd)
}
