package terraform

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/atoz-technology/mantil-backend/internal/assets"
	"github.com/atoz-technology/mantil-backend/internal/mantil"
	"github.com/atoz-technology/mantil-backend/internal/shell"
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

func (t *Terraform) RenderTerraformTemplate(project *mantil.Project) error {
	funcs := template.FuncMap{"join": strings.Join}
	tfTpl, err := assets.Asset("terraform/templates/main.tf")
	if err != nil {
		return err
	}
	tpl := template.Must(template.New("").Funcs(funcs).Parse(string(tfTpl)))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, project); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(t.path, "main.tf"), buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil

}

func (t *Terraform) ApplyForProject(project *mantil.Project, destroy bool) error {
	if err := t.RenderTerraformTemplate(project); err != nil {
		return fmt.Errorf("could not render terraform template for project %s - %v", project.Name, err)
	}
	if err := t.Init(); err != nil {
		return fmt.Errorf("could not init terraform template for project %s - %v", project.Name, err)
	}
	if err := t.Plan(destroy); err != nil {
		return fmt.Errorf("could not plan terraform template for project %s - %v", project.Name, err)
	}
	if err := t.Apply(destroy); err != nil {
		return fmt.Errorf("could not apply terraform template for project %s - %v", project.Name, err)
	}
	return nil
}
