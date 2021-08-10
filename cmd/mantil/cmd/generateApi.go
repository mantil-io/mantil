package cmd

import (
	"fmt"
	"log"

	"github.com/atoz-technology/mantil-cli/internal/generate"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/spf13/cobra"
)

// generateApiCmd represents the generateApi command
var generateApiCmd = &cobra.Command{
	Use:  "generateApi",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		methods, err := cmd.Flags().GetStringSlice("methods")
		if err != nil {
			log.Fatal(err)
		}
		path, err := mantil.FindProjectRoot(".")
		if err != nil {
			log.Fatal(err)
		}
		generateRoot(name, path)
		for _, method := range methods {
			generateMethod(method, name, path)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateApiCmd)
	generateApiCmd.Flags().StringSliceP("methods", "m", nil, "Specify additional function methods, if left empty only the root method will be created.")
}

func generateRoot(name, path string) {
	rootFile := fmt.Sprintf("%s/functions/%s/root.go", path, name)
	if err := generate.GenerateFromTemplate(
		generate.APIRootTemplate,
		&generate.Function{Name: name},
		rootFile,
	); err != nil {
		log.Fatal(err)
	}
}

func generateMethod(name, functionName, path string) {
	methodFile := fmt.Sprintf("%s/functions/%s/%s.go", path, functionName, name)
	if err := generate.GenerateFromTemplate(
		generate.APIMethodTemplate,
		&generate.Method{
			Name:         name,
			FunctionName: functionName,
		},
		methodFile,
	); err != nil {
		log.Fatal(err)
	}
}
