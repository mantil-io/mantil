package cmd

import (
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/spf13/cobra"
)

//var version setup.Version

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mantil",
	Short: "Makes serverless development with Go and AWS Lambda joyful",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// // Execute adds all child commands to the root command and sets flags appropriately.
// // This is called by main.main(). It only needs to happen once to the rootCmd.
// func Execute(v setup.Version) {
// 	version = v
// 	rootCmd.Version = version.String()
// 	cobra.CheckErr(rootCmd.Execute())
// }

func Root() *cobra.Command {
	return rootCmd
}

func init() {
	var verbose, noColor bool
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose log output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "don't use color in log output")

	cobra.OnInitialize(func() {
		if verbose {
			log.EnableDebugLogLevel()
		}
		if noColor {
			log.DisableColor()
		}
	})
}

func getProject() (*config.Project, string) {
	path, err := config.FindProjectRoot(".")
	if err != nil {
		log.Fatal(err)
	}
	p, err := config.LoadProject(path)
	if err != nil {
		log.Error(err)
	}
	return p, path
}
