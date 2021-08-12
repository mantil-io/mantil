package cmd

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/atoz-technology/mantil-cli/internal/generate"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
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
		projectPath, err := mantil.FindProjectRoot(".")
		if err != nil {
			log.Fatal(err)
		}
		importPath := findPackageImportPath(projectPath)
		generateFunctionMain(name, importPath, projectPath)
		generateApi(name, methods, projectPath)
	},
}

func init() {
	generateApiCmd.Flags().StringSliceP("methods", "m", nil, "Specify additional function methods, if left empty only the root method will be created.")
	rootCmd.AddCommand(generateApiCmd)
}

func generateFunctionMain(functionName, importPath, projectPath string) {
	mainFile := fmt.Sprintf("%s/functions/%s/main.go", projectPath, functionName)
	if err := generate.GenerateFromTemplate(
		generate.APIFunctionMainTemplate,
		&generate.Function{
			Name:       functionName,
			ImportPath: findPackageImportPath(projectPath),
		},
		mainFile,
	); err != nil {
		log.Fatal(err)
	}
}

func findPackageImportPath(projectPath string) string {
	modPath := fmt.Sprintf("%s/go.mod", projectPath)
	buf, err := ioutil.ReadFile(modPath)
	if err != nil {
		log.Fatal(err)
	}
	return modfile.ModulePath(buf)
}

func generateApi(functionName string, methods []string, projectPath string) {
	rootFile := fmt.Sprintf("%s/api/%s/root.go", projectPath, functionName)
	if err := generate.GenerateFromTemplate(
		generate.APIRootTemplate,
		&generate.Function{Name: functionName},
		rootFile,
	); err != nil {
		log.Fatal(err)
	}
	for _, method := range methods {
		methodFile := fmt.Sprintf("%s/api/%s/%s.go", projectPath, functionName, method)
		if err := generate.GenerateFromTemplate(
			generate.APIMethodTemplate,
			&generate.Method{
				Name:         method,
				FunctionName: functionName,
			},
			methodFile,
		); err != nil {
			log.Fatal(err)
		}
	}
}
