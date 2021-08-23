package generate

import (
	"fmt"
	"io/ioutil"

	"github.com/mantil-io/mantil-cli/internal/generate"
	"github.com/mantil-io/mantil-cli/internal/mantil"
	"golang.org/x/mod/modfile"
)

func Api(name string, methods []string) error {
	projectPath, err := mantil.FindProjectRoot(".")
	if err != nil {
		return err
	}
	importPath, err := findPackageImportPath(projectPath)
	if err != nil {
		return err
	}
	if err := generateFunctionMain(name, importPath, projectPath); err != nil {
		return err
	}
	return generateApi(name, methods, projectPath)
}

func findPackageImportPath(projectPath string) (string, error) {
	modPath := fmt.Sprintf("%s/go.mod", projectPath)
	buf, err := ioutil.ReadFile(modPath)
	if err != nil {
		return "", err
	}
	return modfile.ModulePath(buf), nil
}

func generateFunctionMain(functionName, importPath, projectPath string) error {
	mainFile := fmt.Sprintf("%s/functions/%s/main.go", projectPath, functionName)
	if err := generate.GenerateFromTemplate(
		generate.APIFunctionMainTemplate,
		&generate.Function{
			Name:       functionName,
			ImportPath: importPath,
		},
		mainFile,
	); err != nil {
		return err
	}
	return nil
}

func generateApi(functionName string, methods []string, projectPath string) error {
	rootFile := fmt.Sprintf("%s/api/%s/root.go", projectPath, functionName)
	if err := generate.GenerateFromTemplate(
		generate.APIRootTemplate,
		&generate.Function{Name: functionName},
		rootFile,
	); err != nil {
		return err
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
			return err
		}
	}
	return nil
}
