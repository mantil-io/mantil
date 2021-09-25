package cmd

import (
	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/cli/commands"
	"github.com/mantil-io/mantil/internal/cli/commands/setup"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/mantil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var verbose bool
var noColor bool
var version setup.Version

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mantil",
	Short: "Makes serverless development with Go and AWS Lambda joyful",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v setup.Version) {
	version = v
	rootCmd.Version = version.String()
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose log output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "don't use color in log output")
}

func initConfig() {
	if verbose {
		log.EnableDebugLogLevel()
	}
	if noColor {
		log.DisableColor()
	}

	viper.AutomaticEnv() // read in environment variables that match
}

func initialiseAWSSDK(projectName, stageName string) *aws.AWS {
	type req struct {
		ProjectName string
		StageName   string
	}
	r := &req{
		ProjectName: projectName,
		StageName:   stageName,
	}
	creds := &commands.Credentials{}
	if err := commands.BackendRequest("security", r, creds, false); err != nil {
		log.Fatal(err)
	}
	awsClient, err := aws.NewWithCredentials(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, creds.Region)
	if err != nil {
		log.Fatal(err)
	}
	return awsClient
}

func getProject() (*mantil.Project, string) {
	path, err := mantil.FindProjectRoot(".")
	if err != nil {
		log.Fatal(err)
	}
	p, err := mantil.LoadProject(path)
	if err != nil {
		log.Error(err)
	}
	return p, path
}
