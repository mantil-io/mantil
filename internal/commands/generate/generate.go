package generate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/mantil-io/mantil-cli/internal/commands/deploy"
	"github.com/mantil-io/mantil-cli/internal/generate"
	"github.com/mantil-io/mantil-cli/internal/log"
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
	return generateApi(projectPath, name, methods)
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
	if fileExists(mainFile) {
		log.Debug("function main already exists, skipping...")
		return nil
	}
	log.Debug("generating function main...")
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

func generateApi(projectPath, functionName string, methods []string) error {
	if err := generateApiDefault(projectPath, functionName); err != nil {
		return err
	}
	if err := generateApiMethods(projectPath, functionName, methods); err != nil {
		return err
	}
	return nil
}

func generateApiDefault(projectPath, functionName string) error {
	defaultFile := path.Join(projectPath, "api", functionName, functionName+".go")
	if fileExists(defaultFile) {
		log.Debug("default method already exists, skipping...")
		return nil
	}
	log.Debug("generating default method...")
	err := generate.GenerateFromTemplate(
		generate.APIDefaultTemplate,
		&generate.Function{Name: functionName},
		defaultFile,
	)
	return err
}

func generateApiMethods(projectPath, functionName string, methods []string) error {
	functionApi := path.Join(projectPath, "api", functionName)
	for _, method := range methods {
		methodFile := path.Join(functionApi, fmt.Sprintf("%s.go", method))
		if fileExists(methodFile) {
			log.Debug("method %s already exists, skipping...", method)
			continue
		}
		log.Debug("generating method %s...", method)
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err == nil {
		return true
	}
	return false
}
