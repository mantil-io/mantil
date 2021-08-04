package cmd

import (
	"log"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/commands/bootstrap"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// bootstrapCmd represents the bootstrap command
var bootstrapCmd = &cobra.Command{
	Use: "bootstrap",
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
		b := bootstrap.New(awsClient)
		destroy, err := cmd.Flags().GetBool("destroy")
		if err != nil {
			log.Fatal(err)
		}
		if err := b.Bootstrap(destroy); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
	bootstrapCmd.Flags().BoolP("destroy", "d", false, "Destroy all resources created by Bootstrap")
}
