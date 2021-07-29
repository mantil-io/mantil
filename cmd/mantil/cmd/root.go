package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/commands"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mantil",
	Short: "A brief description of your application",
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mantil.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".mantil" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mantil")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
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
