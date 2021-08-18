package cmd

import (
	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/commands/setup"
	"github.com/atoz-technology/mantil-cli/internal/log"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setups mantil backend infrastructure in specified AWS account",
	Args:  cobra.NoArgs,
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
		b := setup.New(awsClient)
		destroy, err := cmd.Flags().GetBool("destroy")
		if err != nil {
			log.Fatal(err)
		}
		if err := b.Setup(destroy); err != nil {
			log.Fatal(err)
		}
		if destroy {
			log.Notice("infrastructure successfully destroyed with profile %s", profile)
		} else {
			log.Notice("setup successfully finished with profile %s", profile)
		}
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.Flags().BoolP("destroy", "d", false, "Destroy all resources created by Setup")
}
