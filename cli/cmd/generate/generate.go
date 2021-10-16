package generate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/generate"
	"github.com/mantil-io/mantil/workspace"
	"golang.org/x/mod/modfile"
)

func Api(name string, methods []string) error {
	if !workspace.FunctionNameAvailable(name) {
		return log.Wrap(fmt.Errorf("Could not generate api - name \"%s\" is reserved", name))
	}
	projectPath, err := workspace.FindProjectRoot(".")
	if err != nil {
		return log.Wrap(err)
	}
	importPath, err := findPackageImportPath(projectPath)
	if err != nil {
		return log.Wrap(err)
	}
	if err := generateFunctionMain(name, importPath, projectPath); err != nil {
		return log.Wrap(err)
	}
	if err := generateFunctionGitignore(name, projectPath); err != nil {
		return log.Wrap(err)
	}
	if err := generateFunctionTest(importPath, projectPath, name, methods); err != nil {
		return log.Wrap(err)
	}
	return generateApi(projectPath, name, methods)
}

func findPackageImportPath(projectPath string) (string, error) {
	modPath := filepath.Join(projectPath, "go.mod")
	buf, err := ioutil.ReadFile(modPath)
	if err != nil {
		return "", log.Wrap(err)
	}
	return modfile.ModulePath(buf), nil
}

func generateFunctionMain(functionName, importPath, projectPath string) error {
	functionPath := filepath.Join(deploy.FunctionsDir, functionName)
	root := filepath.Join(projectPath, functionPath)
	mainFile := filepath.Join(root, "main.go")
	if fileExists(mainFile) {
		ui.Info("%s already exists", relativePath(projectPath, mainFile))
		return nil
	}
	ui.Info("Generating %s...", relativePath(projectPath, mainFile))
	if err := generate.GenerateFromTemplate(
		apiFunctionMainTemplate,
		&function{
			Name:       functionName,
			ImportPath: importPath,
		},
		mainFile,
	); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func generateFunctionGitignore(functionName, projectPath string) error {
	functionPath := filepath.Join(deploy.FunctionsDir, functionName)
	gitignoreFile := filepath.Join(projectPath, functionPath, ".gitignore")
	if fileExists(gitignoreFile) {
		ui.Info("%s already exists", relativePath(projectPath, gitignoreFile))
		return nil
	}
	ui.Info("Generating %s...", relativePath(projectPath, gitignoreFile))
	f, err := os.Create(gitignoreFile)
	if err != nil {
		return log.Wrap(err)
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("%s\n", deploy.BinaryName))
	return log.Wrap(err)
}

func generateApi(projectPath, functionName string, methods []string) error {
	if err := generateApiDefault(projectPath, functionName); err != nil {
		return log.Wrap(err)
	}
	if err := generateApiMethods(projectPath, functionName, methods); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func generateApiDefault(projectPath, functionName string) error {
	defaultFile := filepath.Join(projectPath, "api", functionName, fmt.Sprintf("%s.go", functionName))
	if fileExists(defaultFile) {
		ui.Info("%s already exists", relativePath(projectPath, defaultFile))
		return nil
	}
	ui.Info("Generating %s...", relativePath(projectPath, defaultFile))
	if err := generate.GenerateFromTemplate(
		apiDefaultTemplate,
		&function{Name: functionName},
		defaultFile,
	); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func generateApiMethods(projectPath, functionName string, methods []string) error {
	functionApi := filepath.Join(projectPath, "api", functionName)
	for _, m := range methods {
		methodFile := filepath.Join(functionApi, fmt.Sprintf("%s.go", m))
		if fileExists(methodFile) {
			ui.Info("%s already exists", relativePath(projectPath, methodFile))
			continue
		}
		ui.Info("Generating %s...", relativePath(projectPath, methodFile))
		if err := generate.GenerateFromTemplate(
			apiMethodTemplate,
			&method{
				Name:         m,
				FunctionName: functionName,
			},
			methodFile,
		); err != nil {
			return log.Wrap(err)
		}
	}
	return nil
}

func generateFunctionTest(importPath, projectPath, functionName string, methods []string) error {
	if err := generateApiTestInit(projectPath); err != nil {
		return log.Wrap(err)
	}
	if err := generateApiTest(importPath, projectPath, functionName, methods); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func generateApiTestInit(projectPath string) error {
	initTest := filepath.Join(projectPath, "test", "init.go")
	if fileExists(initTest) {
		ui.Info("%s already exists", relativePath(projectPath, initTest))
		return nil
	}
	ui.Info("Generating %s...", relativePath(projectPath, initTest))
	if err := generate.GenerateFile(
		apiFunctionTestInit,
		initTest,
	); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func generateApiTest(importPath, projectPath, functionName string, methods []string) error {
	apiTest := filepath.Join(projectPath, "test", fmt.Sprintf("%s_test.go", strings.ToLower(functionName)))
	if fileExists(apiTest) {
		ui.Info("%s already exists", relativePath(projectPath, apiTest))
		return nil
	}
	ui.Info("Generating %s...", relativePath(projectPath, apiTest))
	if err := generate.GenerateFromTemplate(
		apiFunctionTestTemplate,
		&test{
			Name:       functionName,
			ImportPath: importPath,
			Methods:    methods,
		},
		apiTest,
	); err != nil {
		return log.Wrap(err)
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
