package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Creates infrastructure and deploys updates to lambda functions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("deploy")
	},
}

func init() {
	rootCmd.AddCommand(deployCmd)
}
