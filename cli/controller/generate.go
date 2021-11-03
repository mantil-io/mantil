package controller

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"golang.org/x/mod/modfile"
)

type GenerateApiArgs struct {
	Name    string
	Methods []string
}

func GenerateApi(a GenerateApiArgs) error {
	if !domain.FunctionNameAvailable(a.Name) {
		return log.Wrap(fmt.Errorf("Could not generate api - name \"%s\" is reserved", a.Name))
	}
	if err := domain.ValidateName(a.Name); err != nil {
		return log.Wrap(err)
	}

	msg := fmt.Sprintf("Generating function %s", a.Name)
	if len(a.Methods) > 0 {
		msg = fmt.Sprintf("%s with additional methods %s", msg, strings.Join(a.Methods, ","))
	}
	ui.Info("%s\n", msg)

	projectPath, err := domain.FindProjectRoot(".")
	if err != nil {
		return log.Wrap(err)
	}
	importPath, err := findPackageImportPath(projectPath)
	if err != nil {
		return log.Wrap(err)
	}
	if err := generateFunctionMain(a.Name, importPath, projectPath); err != nil {
		return log.Wrap(err)
	}
	if err := generateFunctionGitignore(a.Name, projectPath); err != nil {
		return log.Wrap(err)
	}
	if err := generateFunctionTest(importPath, projectPath, a.Name, a.Methods); err != nil {
		return log.Wrap(err)
	}
	return generateApi(projectPath, a.Name, a.Methods)
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
	functionPath := filepath.Join(FunctionsDir, functionName)
	root := filepath.Join(projectPath, functionPath)
	mainFile := filepath.Join(root, "main.go")
	if fileExists(mainFile) {
		ui.Info("%s already exists", relativePath(projectPath, mainFile))
		return nil
	}
	ui.Info("Generating %s...", relativePath(projectPath, mainFile))
	if err := generateFromTemplate(
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
	functionPath := filepath.Join(FunctionsDir, functionName)
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
	_, err = f.WriteString(fmt.Sprintf("%s\n", BinaryName))
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
	if err := generateFromTemplate(
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
		if err := generateFromTemplate(
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
	if err := generateFile(
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
	if err := generateFromTemplate(
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

func generateFromTemplate(tplDef string, data interface{}, outPath string) error {
	out, err := renderTemplate(tplDef, data)
	if err != nil {
		return err
	}
	out, err = goFmt(string(out))
	if err != nil {
		return err
	}
	return saveFile(out, outPath)
}

func generateFile(content string, outPath string) error {
	out, err := goFmt(content)
	if err != nil {
		return err
	}
	return saveFile(out, outPath)
}

func goFmt(in string) ([]byte, error) {
	cmd := exec.Command("gofmt")
	cmd.Stdin = strings.NewReader(in)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func saveFile(in []byte, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, in, 0644); err != nil {
		return err
	}
	return nil
}
