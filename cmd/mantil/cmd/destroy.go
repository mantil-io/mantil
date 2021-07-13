package cmd

import (
	"log"

	"github.com/atoz-technology/mantil-cli/internal/commands/destroy"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:  "destroy",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		d := destroy.New()
		if err := d.DestroyInfrastructure(name); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
