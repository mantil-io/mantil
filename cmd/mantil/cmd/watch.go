package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil-cli/internal/commands/deploy"
	"github.com/mantil-io/mantil-cli/internal/commands/invoke"
	"github.com/mantil-io/mantil-cli/internal/commands/watch"
	"github.com/mantil-io/mantil-cli/internal/log"
	"github.com/mantil-io/mantil.go/pkg/shell"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for file changes and automatically deploy functions",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		config, path, token := localData()
		p := fetchProject(config.Name, token)

		method := cmd.Flag("method").Value.String()
		test, _ := cmd.Flags().GetBool("test")
		data := cmd.Flag("data").Value.String()

		if method != "" && config.ApiURL == "" {
			log.Fatalf("api URL for the project does not exist")
		}
		endpoint := fmt.Sprintf("%s/%s", config.ApiURL, method)
		aws := initialiseAWSSDK(config.Name, token)

		d, err := deploy.New(p, config, aws, path, token)
		if err != nil {
			log.Fatal(err)
		}

		watch.Start(path, func() {
			log.Info("\nchanges detected - starting deploy")
			if err := d.Deploy(); err != nil {
				log.Fatal(err)
			}
			if !d.HasUpdates() {
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
	rootCmd.AddCommand(watchCmd)
}
