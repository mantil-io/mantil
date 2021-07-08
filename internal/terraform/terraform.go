package terraform

import (
	"fmt"
	"os"
	"strings"

	"github.com/atoz-technology/mantil-cli/pkg/shell"
)

type Terraform struct {
	path string
}

func New(path string) *Terraform {
	return &Terraform{
		path: path,
	}
}

func (t *Terraform) Plan() error {
	return shell.Exec([]string{"terraform", "plan", "-no-color", "-input=false", "-out=tfplan"}, t.path)
}

func (t *Terraform) Apply() error {
	return shell.Exec([]string{"terraform", "apply", "-no-color", "-input=false", "tfplan"}, t.path)
}

func (t *Terraform) Init() error {
	if _, err := os.Stat(t.path + "/.terraform"); os.IsNotExist(err) { // only if .terraform folder not found
		return shell.Exec([]string{"terraform", "init", "-no-color", "-input=false"}, t.path)
	}
	return nil
}

func (t *Terraform) Output(key string) (string, error) {
	val, err := shell.Output([]string{"terraform", "output", "--json", key}, t.path)
	if err != nil {
		return "", err
	}
	if strings.Contains(val, "No outputs found") {
		return "", fmt.Errorf("can't read entrypoint")
	}
	return val, nil
}
