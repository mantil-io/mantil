package generate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/commands/deploy"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/generate"
	"golang.org/x/mod/modfile"
)

func Api(name string, methods []string) error {
	if !config.FunctionNameAvailable(name) {
		return fmt.Errorf("could not generate api - name \"%s\" is reserved", name)
	}
	projectPath, err := config.FindProjectRoot(".")
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
	if err := generateFunctionTest(importPath, projectPath, name, methods); err != nil {
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
	functionPath := filepath.Join(deploy.FunctionsDir, functionName)
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
	functionPath := filepath.Join(deploy.FunctionsDir, functionName)
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

func generateFunctionTest(importPath, projectPath, functionName string, methods []string) error {
	if err := generateApiTestInit(projectPath); err != nil {
		return err
	}
	if err := generateApiTest(importPath, projectPath, functionName, methods); err != nil {
		return err
	}
	return nil
}

func generateApiTestInit(projectPath string) error {
	initTest := filepath.Join(projectPath, "test", "init.go")
	if fileExists(initTest) {
		log.Info("%s already exists", relativePath(projectPath, initTest))
		return nil
	}
	log.Info("generating %s", relativePath(projectPath, initTest))
	if err := generate.GenerateFile(
		generate.APIFunctionTestInit,
		initTest,
	); err != nil {
		return err
	}
	return nil
}

func generateApiTest(importPath, projectPath, functionName string, methods []string) error {
	apiTest := filepath.Join(projectPath, "test", fmt.Sprintf("%s_test.go", strings.ToLower(functionName)))
	if fileExists(apiTest) {
		log.Info("%s already exists", relativePath(projectPath, apiTest))
		return nil
	}
	log.Info("generating %s", relativePath(projectPath, apiTest))
	if err := generate.GenerateFromTemplate(
		generate.APIFunctionTestTemplate,
		&generate.Test{
			Name:       functionName,
			ImportPath: importPath,
			Methods:    methods,
		},
		apiTest,
	); err != nil {
		return err
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
