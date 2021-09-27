package cmd

import (
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/spf13/cobra"
)

func init() {
	addCommandEnv()
	addCommandNew()
}

func addCommandEnv() {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Show project environment variables",
		Long: `Show project environment variables

You can set environment variables in terminal with:
$ eval $(mantil env)
`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			ec := initEnv(cmd, args)
			if err := ec.run(); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().BoolP("url", "u", false, "show only project api url")
	cmd.Flags().StringP("stage", "s", config.DefaultStageName, "stage name")
	rootCmd.AddCommand(cmd)
}

func initEnv(cmd *cobra.Command, args []string) *envCmd {
	url, _ := cmd.Flags().GetBool("url")
	stageName, _ := cmd.Flags().GetString("stage")
	return &envCmd{
		url:       url,
		stageName: stageName,
	}
}
func addCommandNew() {
	cmd := &cobra.Command{
		Use:   "new <project>",
		Short: "Initializes a new Mantil project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			nc := initNew(cmd, args)
			if err := nc.run(); err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().String("from", "", "name of the template or URL of the repository that will be used as one")
	cmd.Flags().String("module-name", "", "replace module name and import paths")
	rootCmd.AddCommand(cmd)
}

func initNew(cmd *cobra.Command, args []string) *newCmd {
	projectName := args[0]
	repo := cmd.Flag("from").Value.String()
	moduleName := cmd.Flag("module-name").Value.String()
	if moduleName == "" {
		moduleName = projectName
	}
	return &newCmd{
		name:       projectName,
		repo:       repo,
		moduleName: moduleName,
	}
}
