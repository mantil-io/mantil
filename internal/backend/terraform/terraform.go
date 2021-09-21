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

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/backend/assets"
	"github.com/mantil-io/mantil/internal/backend/log"
	"github.com/mantil-io/mantil/internal/mantil"
	"github.com/mantil-io/mantil/internal/shell"
)

type Terraform struct {
	path string
}

func New(projectName string) (*Terraform, error) {
	path, err := LambdaProjectDir(projectName)
	if err != nil {
		return nil, err
	}
	return &Terraform{
		path: path,
	}, nil
}

func LambdaProjectDir(projectName string) (string, error) {
	dir := fmt.Sprintf("/tmp/%s", projectName)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}
	return dir, nil
}

func (t *Terraform) init() error {
	if _, err := os.Stat(t.path + "/.terraform"); os.IsNotExist(err) { // only if .terraform folder not found
		return shell.Exec(shell.ExecOptions{
			Args:    []string{"terraform", "init", "-no-color", "-input=false", "-migrate-state"},
			WorkDir: t.path,
			Logger:  t.shellOutput(),
		})
	}
	return nil
}

func (t *Terraform) plan(destroy bool) error {
	args := []string{"terraform", "plan", "-no-color", "-input=false", "-out=tfplan"}
	if destroy {
		args = append(args, "-destroy")
	}
	return shell.Exec(shell.ExecOptions{
		Args:    args,
		WorkDir: t.path,
		Logger:  t.shellOutput(),
	})
}

func (t *Terraform) apply(destroy bool) error {
	args := []string{"terraform", "apply", "-no-color", "-input=false"}
	if destroy {
		args = append(args, "-destroy")
	}
	args = append(args, "tfplan")
	return shell.Exec(shell.ExecOptions{
		Args:    args,
		WorkDir: t.path,
		Logger:  t.shellOutput(),
	})
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
	if err := t.init(); err != nil {
		return "", err
	}
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

func (t *Terraform) ApplyForProject(project *mantil.Project, aws *aws.AWS, destroy bool) error {
	data := struct {
		Name           string
		Bucket         string
		BucketPrefix   string
		Functions      []mantil.Function
		StaticWebsites []mantil.StaticWebsite
		Region         string
	}{
		project.Name,
		project.Bucket,
		project.BucketPrefix,
		project.Functions,
		project.StaticWebsites,
		aws.Region(),
	}
	if err := t.RenderTerraformTemplate("terraform/templates/project.tf", &data); err != nil {
		return fmt.Errorf("could not render terraform template for project %s - %v", project.Name, err)
	}
	if err := t.Apply(destroy); err != nil {
		return fmt.Errorf("could not apply terraform template for project - %v", err)
	}
	return nil
}

func (t *Terraform) RenderSetupTemplate(bucket string, aws *aws.AWS) error {
	data := struct {
		Bucket       string
		BucketPrefix string
		Region       string
	}{
		bucket,
		mantil.SetupBucketPrefix(),
		aws.Region(),
	}
	if err := t.RenderTerraformTemplate("terraform/templates/setup.tf", &data); err != nil {
		return fmt.Errorf("could not render terraform template for setup - %v", err)
	}
	return nil
}

func (t *Terraform) Apply(destroy bool) error {
	if err := t.init(); err != nil {
		return err
	}
	if err := t.plan(destroy); err != nil {
		return err
	}
	if err := t.apply(destroy); err != nil {
		return err
	}
	return nil
}

func (t *Terraform) Cleanup() {
	if err := os.RemoveAll(t.path); err != nil {
		log.Error(err)
	}
}
