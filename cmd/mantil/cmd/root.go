package cmd

import (
	"github.com/mantil-io/mantil-cli/internal/aws"
	"github.com/mantil-io/mantil-cli/internal/commands"
	"github.com/mantil-io/mantil-cli/internal/log"
	"github.com/mantil-io/mantil-cli/internal/mantil"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var verbose bool
var noColor bool

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
func Execute() {
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

func initialiseAWSSDK(projectName, token string) (*aws.AWS, error) {
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
		return nil, err
	}
	awsClient, err := aws.New(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)
	if err != nil {
		return nil, err
	}
	return awsClient, nil
}

func findProject(args []string) (*mantil.Project, *mantil.LocalProjectConfig, string, string) {
	initPath := "."
	if len(args) >= 1 {
		initPath = args[0]
	}
	path, err := mantil.FindProjectRoot(initPath)
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
	project, err := fetchProject(config.Name, token)
	if err != nil {
		log.Fatal(err)
	}
	return project, config, path, token
}

func fetchProject(projectName, token string) (*mantil.Project, error) {
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
		return nil, err
	}
	return rsp.Project, nil
}
