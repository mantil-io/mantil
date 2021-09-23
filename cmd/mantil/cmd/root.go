package cmd

import (
	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/cli/commands"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/mantil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var verbose bool
var noColor bool
var version Version

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mantil",
	Short: "Makes serverless development with Go and AWS Lambda joyful",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

type Version struct {
	Commit        string
	Tag           string
	Dirty         bool
	Version       string
	FunctionsPath string
}

func (v *Version) String() string {
	return v.Version
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v Version) {
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

func initialiseAWSSDK(projectName, token string) *aws.AWS {
	type req struct {
		ProjectName string
		Token       string
	}
	r := &req{
		ProjectName: projectName,
		Token:       token,
	}
	creds := &commands.Credentials{}
	if err := commands.BackendRequest("security", r, creds); err != nil {
		log.Fatal(err)
	}
	awsClient, err := aws.NewWithCredentials(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, creds.Region)
	if err != nil {
		log.Fatal(err)
	}
	return awsClient
}

func localData() (*mantil.LocalProjectConfig, string, string) {
	path, err := mantil.FindProjectRoot(".")
	if err != nil {
		log.Fatal(err)
	}
	config, err := mantil.LoadLocalConfig(path)
	if err != nil {
		log.Fatal(err)
	}
	token, err := mantil.ReadToken(config.Name)
	if err != nil {
		log.Fatal(err)
	}
	return config, path, token
}

func fetchProject(projectName, token string) *mantil.Project {
	type req struct {
		ProjectName string
		Token       string
	}
	r := &req{
		ProjectName: projectName,
		Token:       token,
	}

	type resp struct {
		Project *mantil.Project
	}
	rsp := &resp{}
	if err := commands.BackendRequest("data", r, rsp); err != nil {
		log.Fatal(err)
	}
	return rsp.Project
}
