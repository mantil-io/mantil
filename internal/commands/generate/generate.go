package generate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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
	if err := generateFunctionGitignore(name, projectPath); err != nil {
		return err
	}
	return generateApi(projectPath, name, methods)
}

func findPackageImportPath(projectPath string) (string, error) {
	modPath := filepath.Join(projectPath, "go.mod")
	buf, err := ioutil.ReadFile(modPath)
	if err != nil {
		return "", err
	}
	return modfile.ModulePath(buf), nil
}

func generateFunctionMain(functionName, importPath, projectPath string) error {
	functionPath := filepath.Join("functions", functionName)
	root := filepath.Join(projectPath, functionPath)
	mainFile := filepath.Join(root, "main.go")
	if fileExists(mainFile) {
		log.Info("%s already exists", relativePath(projectPath, mainFile))
		return nil
	}
	log.Info("generating %s", relativePath(projectPath, mainFile))
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

func generateFunctionGitignore(functionName, projectPath string) error {
	functionPath := filepath.Join("functions", functionName)
	gitignoreFile := filepath.Join(projectPath, functionPath, ".gitignore")
	if fileExists(gitignoreFile) {
		log.Info("%s already exists", relativePath(projectPath, gitignoreFile))
		return nil
	}
	log.Info("generating %s", relativePath(projectPath, gitignoreFile))
	f, err := os.Create(gitignoreFile)
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
	defaultFile := filepath.Join(projectPath, "api", functionName, fmt.Sprintf("%s.go", functionName))
	if fileExists(defaultFile) {
		log.Info("%s already exists", relativePath(projectPath, defaultFile))
		return nil
	}
	log.Info("generating %s", relativePath(projectPath, defaultFile))
	err := generate.GenerateFromTemplate(
		generate.APIDefaultTemplate,
		&generate.Function{Name: functionName},
		defaultFile,
	)
	return err
}

func generateApiMethods(projectPath, functionName string, methods []string) error {
	functionApi := filepath.Join(projectPath, "api", functionName)
	for _, method := range methods {
		methodFile := filepath.Join(functionApi, fmt.Sprintf("%s.go", method))
		if fileExists(methodFile) {
			log.Info("%s already exists", relativePath(projectPath, methodFile))
			continue
		}
		log.Info("generating %s", relativePath(projectPath, methodFile))
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

func relativePath(basePath, targPath string) string {
	rel, err := filepath.Rel(basePath, targPath)
	if err != nil {
		return targPath
	}
	return rel
}
