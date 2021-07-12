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

func (t *Terraform) Plan(destroy bool) error {
	args := []string{"terraform", "plan", "-no-color", "-input=false", "-out=tfplan"}
	if destroy {
		args = append(args, "-destroy")
	}
	return shell.Exec(args, t.path)
}

func (t *Terraform) Apply(destroy bool) error {
	args := []string{"terraform", "apply", "-no-color", "-input=false"}
	if destroy {
		args = append(args, "-destroy")
	}
	args = append(args, "tfplan")
	return shell.Exec(args, t.path)
}

func (t *Terraform) Init() error {
	if _, err := os.Stat(t.path + "/.terraform"); os.IsNotExist(err) { // only if .terraform folder not found
		return shell.Exec([]string{"terraform", "init", "-no-color", "-input=false", "-migrate-state"}, t.path)
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
