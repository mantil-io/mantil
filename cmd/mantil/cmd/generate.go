package cmd

import (
	"github.com/mantil-io/mantil/cmd/mantil/commands/generate"
	"github.com/mantil-io/mantil/cmd/mantil/log"
	"github.com/spf13/cobra"
)

// generateCmd represents generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Automatically generate code in the project",
}

var generateApiCmd = &cobra.Command{
	Use:   "api <function>",
	Short: "Generate Go code for new api",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		methods, err := cmd.Flags().GetStringSlice("methods")
		if err != nil {
			log.Fatal(err)
		}
		if err := generate.Api(name, methods); err != nil {
			log.Fatal(err)
		}
		log.Notice("successfuly generated api %s", name)
	},
}

func init() {
	generateApiCmd.Flags().StringSliceP("methods", "m", nil, "additional function methods, if left empty only the Default method will be created")
	generateCmd.AddCommand(generateApiCmd)
	rootCmd.AddCommand(generateCmd)
}
