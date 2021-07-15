package cmd

import (
	"log"

	"github.com/atoz-technology/mantil-cli/internal/commands/destroy"
	"github.com/spf13/cobra"
)

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:  "destroy",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		d, err := destroy.New(findProject(args))
		if err != nil {
			log.Fatal(err)
		}
		if err := d.Destroy(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
