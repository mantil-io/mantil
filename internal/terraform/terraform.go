package terraform

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/mantil-io/mantil-backend/internal/assets"
	"github.com/mantil-io/mantil-backend/internal/log"
	"github.com/mantil-io/mantil-backend/internal/mantil"
	"github.com/mantil-io/mantil-backend/internal/shell"
)

type Terraform struct {
	path string
}

func New(path string) *Terraform {
	return &Terraform{
		path: path,
	}
}

func (t *Terraform) Init() error {
	if _, err := os.Stat(t.path + "/.terraform"); os.IsNotExist(err) { // only if .terraform folder not found
		return shell.ExecWithOutput([]string{"terraform", "init", "-no-color", "-input=false", "-migrate-state"}, t.path, t.shellOutput())
	}
	return nil
}

func (t *Terraform) Plan(destroy bool) error {
	args := []string{"terraform", "plan", "-no-color", "-input=false", "-out=tfplan"}
	if destroy {
		args = append(args, "-destroy")
	}
	return shell.ExecWithOutput(args, t.path, t.shellOutput())
}

func (t *Terraform) Apply(destroy bool) error {
	args := []string{"terraform", "apply", "-no-color", "-input=false"}
	if destroy {
		args = append(args, "-destroy")
	}
	args = append(args, "tfplan")
	return shell.ExecWithOutput(args, t.path, t.shellOutput())
}

func (t *Terraform) shellOutput() func(string, ...interface{}) {
	var (
		isError                  = false
		terraformCreatedRegexp   = regexp.MustCompile(`\w\.(.*): Creation complete after (\w*) `)
		terraformDestroyedRegexp = regexp.MustCompile(`\w\.(.*): Destruction complete after (\w*)`)
		terraformCompleteRegexp  = regexp.MustCompile(`Apply complete! Resources: (\w*) added, (\w*) changed, (\w*) destroyed.`)
	)
	terraformCreated := func(line string) string {
		match := terraformCreatedRegexp.FindStringSubmatch(line)
		if len(match) == 3 {
			return fmt.Sprintf("Created %s in %s", match[1], match[2])
		}
		return ""
	}

	terraformDestroyed := func(line string) string {
		match := terraformDestroyedRegexp.FindStringSubmatch(line)
		if len(match) == 3 {
			return fmt.Sprintf("Destroyed %s in %s", match[1], match[2])
		}
		return ""
	}

	terraformComplete := func(line string) string {
		match := terraformCompleteRegexp.FindStringSubmatch(line)
		if len(match) == 4 {
			return fmt.Sprintf("Resources: %s added, %s changed, %s destroyed", match[1], match[2], match[3])
		}
		return ""
	}

	output := func(format string, v ...interface{}) {
		msg := fmt.Sprintf(format, v...)

		// if error line was encountered print out the rest of the lines as errors since they're useful for debugging
		if isError {
			log.Errorf(msg)
			return
		}

		if l := terraformCreated(msg); l != "" {
			log.Info(l)
		} else if l := terraformDestroyed(msg); l != "" {
			log.Info(l)
		} else if l := terraformComplete(msg); l != "" {
			log.Info(l)
		} else if strings.HasPrefix(msg, "Error:") {
			log.Errorf(msg)
			isError = true
		}
	}
	return output
}

func (t *Terraform) Output(key string, raw bool) (string, error) {
	var args []string
	if raw {
		args = []string{"terraform", "output", "-raw", key}
	} else {
		args = []string{"terraform", "output", "-json", key}
	}
	val, err := shell.Output(args, t.path)
	if err != nil {
		return "", err
	}
	if strings.Contains(val, "No outputs found") {
		return "", fmt.Errorf("can't read entrypoint")
	}
	return val, nil
}

func (t *Terraform) RenderTerraformTemplate(templatePath string, data interface{}) error {
	funcs := template.FuncMap{"join": strings.Join}
	tfTpl, err := assets.Asset(templatePath)
	if err != nil {
		return err
	}
	tpl := template.Must(template.New("").Funcs(funcs).Parse(string(tfTpl)))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(t.path, "main.tf"), buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil

}

func (t *Terraform) ApplyForProject(project *mantil.Project, destroy bool) error {
	if err := t.RenderTerraformTemplate("terraform/templates/project.tf", project); err != nil {
		return fmt.Errorf("could not render terraform template for project %s - %v", project.Name, err)
	}
	if err := t.apply(destroy); err != nil {
		return fmt.Errorf("could not apply terraform template for project - %v", err)
	}
	return nil
}

func (t *Terraform) ApplyForSetup(bucket string, destroy bool) error {
	type data struct {
		Bucket string
	}
	if err := t.RenderTerraformTemplate("terraform/templates/setup.tf", &data{bucket}); err != nil {
		return fmt.Errorf("could not render terraform template for setup - %v", err)
	}
	if err := t.apply(destroy); err != nil {
		return fmt.Errorf("could not apply setup terraform template - %v", err)
	}
	return nil
}

func (t *Terraform) apply(destroy bool) error {
	if err := t.Init(); err != nil {
		return err
	}
	if err := t.Plan(destroy); err != nil {
		return err
	}
	if err := t.Apply(destroy); err != nil {
		return err
	}
	return nil
}
