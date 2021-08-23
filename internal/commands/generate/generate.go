package generate

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/mantil-io/mantil-cli/internal/commands/deploy"
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
	root := path.Join(projectPath, "functions", functionName)
	mainFile := path.Join(root, "main.go")
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
	if err := createFunctionGitignore(root); err != nil {
		return err
	}
	return nil
}

func createFunctionGitignore(root string) error {
	f, err := os.Create(path.Join(root, ".gitignore"))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%s\n", deploy.BinaryName))
	return err
}

func generateApi(functionName string, methods []string, projectPath string) error {
	defaultFile := fmt.Sprintf("%s/api/%s/default.go", projectPath, functionName)
	if err := generate.GenerateFromTemplate(
		generate.APIDefaultTemplate,
		&generate.Function{Name: functionName},
		defaultFile,
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
