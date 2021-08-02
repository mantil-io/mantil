package cmd

import (
	"fmt"
	"log"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use: "new",
	Run: func(cmd *cobra.Command, args []string) {
		profiles, err := aws.ListProfiles()
		if err != nil {
			log.Fatal(err)
		}
		prompt := promptui.Select{
			Label: "Select AWS profile",
			Items: profiles,
		}
		_, profile, err := prompt.Run()
		if err != nil {
			log.Fatal(err)
		}
		awsClient, err := aws.NewFromProfile(profile)
		if err != nil {
			log.Fatal(err)
		}
		creds, _ := awsClient.Credentials()
		fmt.Println(creds.AccessKeyID)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
