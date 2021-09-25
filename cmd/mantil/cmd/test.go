package cmd

import (
	"github.com/mantil-io/mantil.go/pkg/shell"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/config"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run project integration tests",
	Long: `Run project integration tests

Project integration tests are pure Go test in [project-root]/test folder.
Mantil sets MANTIL_API_URL environment variable to point to the current
project api url and runs tests with 'go test -v'.
`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		p, path := getProject()
		run := cmd.Flag("run").Value.String()
		shellArgs := []string{"go", "test", "-v"}
		if run != "" {
			shellArgs = append(shellArgs, "--run", run)
		}
		stageName, _ := cmd.Flags().GetString("stage")
		err := shell.Exec(shell.ExecOptions{
			Env:          []string{"MANTIL_API_URL=" + p.RestEndpoint(stageName)},
			Args:         shellArgs,
			WorkDir:      path + "/test",
			Logger:       log.Info,
			ShowShellCmd: false,
		})
		if err != nil {
			log.Error(err)
		}
	},
}

func init() {
	testCmd.Flags().StringP("run", "r", "", "run only tests with this pattern in name")
	testCmd.Flags().StringP("stage", "s", config.DefaultStageName, "stage name")
	rootCmd.AddCommand(testCmd)
}
