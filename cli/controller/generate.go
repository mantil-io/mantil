package controller

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/imports"
)

type ApiNewError struct {
	Api string
}

func (e ApiNewError) Error() string {
	return fmt.Sprintf("function New for api %s doesn't have proper type", e.Api)
}

type GenerateApiArgs struct {
	Name    string
	Methods []string
}

func GenerateApi(a GenerateApiArgs) error {
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
	out, err = formatAndAdjustImports(string(out))
	if err != nil {
		return err
	}
	return saveFile(out, outPath)
}

func generateFile(content string, outPath string) error {
	out, err := formatAndAdjustImports(content)
	if err != nil {
		return err
	}
	return saveFile(out, outPath)
}

func formatAndAdjustImports(in string) ([]byte, error) {
	out, err := imports.Process("", []byte(in), nil)
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

func generateMain(api, apiDir, destination string) error {
	if err := isNewValid(api, apiDir); err != nil {
		return log.Wrap(err)
	}
	projectPath, err := domain.FindProjectRoot(".")
	if err != nil {
		return log.Wrap(err)
	}
	importPath, err := findPackageImportPath(projectPath)
	if err != nil {
		return log.Wrap(err)
	}
	if err := generateFromTemplate(
		apiFunctionMainTemplate,
		&function{
			Name:       api,
			ImportPath: importPath,
		},
		destination,
	); err != nil {
		return log.Wrap(err)
	}
	return nil
}

// isNewValid checks whether function New in api is of proper type
// function should have no parameters and only one return value - struct or pointer to the struct
func isNewValid(api, dir string) error {
	pkgs, err := parser.ParseDir(token.NewFileSet(), dir, nil, parser.AllErrors)
	if err != nil {
		return log.Wrap(err)
	}
	pkg, ok := pkgs[api]
	if !ok {
		return log.Wrapf("package %s doesn't exist in folder %s", api, dir)
	}
	for _, v := range pkg.Files {
		for _, o := range v.Scope.Objects {
			if o.Name == "New" && o.Kind.String() == "func" {
				decl, ok := o.Decl.(*ast.FuncDecl)
				if !ok {
					return log.Wrap(&ApiNewError{api})
				}
				// is not a function
				if decl.Recv != nil {
					continue
				}
				// has no parameters
				if len(decl.Type.Params.List) > 0 {
					return log.Wrap(&ApiNewError{api})
				}
				rl := decl.Type.Results.List
				// has only one return value which is either struct or pointer to struct
				if len(rl) > 1 {
					return log.Wrap(&ApiNewError{api})
				}
				var idExpr ast.Expr
				expr, ok := rl[0].Type.(*ast.StarExpr)
				if ok {
					idExpr = expr.X
				} else {
					idExpr = rl[0].Type
				}
				ident, ok := idExpr.(*ast.Ident)
				if ok && ident.Obj != nil {
					ft, ok := ident.Obj.Decl.(*ast.TypeSpec)
					if ok {
						_, ok := ft.Type.(*ast.StructType)
						if ok {
							return nil
						}
					}
				}
				return log.Wrap(&ApiNewError{api})
			}
		}
	}
	return log.Wrap(&ApiNewError{api})
}
