package terraform

import (
	"bytes"
	"embed"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/shell"
	"github.com/mantil-io/mantil/workspace"
)

//go:embed modules/* templates/*
var fs embed.FS

const (
	rootPath            = "/tmp/mantil"
	modulesDir          = "modules"
	templatesDir        = "templates"
	destroyTemplateName = "destroy.tf"
	setupTemplateName   = "setup.tf"
	projectTemplateName = "project.tf"
	mainTf              = "main.tf"
	setupBucketPrefix   = "setup"
	createDir           = "create"
	destroyDir          = "destroy"
)

type Terraform struct {
	path        string
	createPath  string
	destroyPath string
}

type SetupTemplateData struct {
	Bucket          string
	BucketPrefix    string
	FunctionsBucket string
	FunctionsPath   string
	Region          string
	PublicKey       string
}

// Prepare setup templates
func Setup(data SetupTemplateData) (*Terraform, error) {
	if err := extractModules(); err != nil {
		return nil, err
	}
	return renderSetup(data)
}

type ProjectTemplateData struct {
	Name                   string
	Bucket                 string
	BucketPrefix           string
	Functions              []*workspace.Function
	Public                 []*workspace.PublicSite
	Region                 string
	Stage                  string
	RuntimeFunctionsBucket string
	RuntimeFunctionsPath   string
	GlobalEnv              map[string]string
	// TODO: uskladi nazivlje u struct gore i ovdje FunctionsBucket i Path
}

// Prepare project templates
func Project(data ProjectTemplateData) (*Terraform, error) {
	if err := extractModules(); err != nil {
		return nil, err
	}
	return renderProject(data)
}

// Create or apply changes to the infrastructure
func (t *Terraform) Create() error {
	t.path = t.createPath
	// retry on ConflictException
	for {
		err := t.initPlanApply(false)
		if err == nil || err != conflictException {
			return err
		}
	}
}

// Destroy all infrastructure resources
func (t *Terraform) Destroy() error {
	t.path = t.destroyPath
	return t.initPlanApply(true)
}

// Output reads terrraform output variable value
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
	val, err := shell.Output(t.shellExecOpts("", args))
	if err != nil {
		return "", err
	}
	if strings.Contains(val, "No outputs found") {
		return "", fmt.Errorf("no outputs found")
	}
	return val, nil
}

// path to create/main.tf
func (t *Terraform) CreateTf() string {
	return path.Join(t.createPath, mainTf)
}

// path to destsroy/main.tf
func (t *Terraform) DestroyTf() string {
	return path.Join(t.destroyPath, mainTf)
}

func (t *Terraform) initPlanApply(destroy bool) error {
	if err := t.init(); err != nil {
		return err
	}
	if err := t.plan(destroy); err != nil {
		return err
	}
	if err := t.apply(destroy); err != nil {
		return err
	}
	if !destroy {
		return t.output()
	}
	return nil
}

func (t *Terraform) init() error {
	if _, err := os.Stat(t.path + "/.terraform"); os.IsNotExist(err) { // only if .terraform folder not found
		args := []string{"terraform", "init", "-no-color", "-input=false", "-migrate-state"}
		return t.shellExec(args)
	}
	return nil
}

func (t *Terraform) plan(destroy bool) error {
	args := []string{"terraform", "plan", "-no-color", "-input=false", "-out=tfplan"}
	if destroy {
		args = append(args, "-destroy")
	}
	return t.shellExec(args)
}

func (t *Terraform) apply(destroy bool) error {
	args := []string{"terraform", "apply", "-no-color", "-input=false"}
	if destroy {
		args = append(args, "-destroy")
	}
	args = append(args, "tfplan")
	opt := t.shellExecOpts(logPrefix, args)
	opt.ErrorsMap = map[string]error{
		"ConflictException: Unable to complete operation due to concurrent modification. Please try again later.": conflictException,
	}
	return shell.Exec(opt)
}

func (t *Terraform) output() error {
	args := []string{"terraform", "output", "-no-color"}
	return shell.Exec(t.shellExecOpts(outputLogPrefix, args))
}

/////////////// shell exec

var conflictException = fmt.Errorf("ConflictException")

func (t *Terraform) shellExecOpts(logPrefix string, args []string) shell.ExecOptions {
	opt := shell.ExecOptions{
		Args:         args,
		WorkDir:      t.path,
		ShowShellCmd: true,
		Logger: func(format string, v ...interface{}) {
			format = logPrefix + format
			stdlog.Printf(format, v...)
		},
	}
	if p := aws.TestProfile(); p != "" {
		opt.Env = []string{"AWS_PROFILE=" + p}
	}
	return opt
}

func (t *Terraform) shellExec(args []string) error {
	return shell.Exec(t.shellExecOpts(logPrefix, args))
}

/////////////// rendering templates

func renderProject(data ProjectTemplateData) (*Terraform, error) {
	stageDir := fmt.Sprintf("%s-%s", data.Name, data.Stage)
	t := &Terraform{
		createPath:  path.Join(rootPath, stageDir, createDir),
		destroyPath: path.Join(rootPath, stageDir, destroyDir),
	}
	if err := t.render(projectTemplateName, t.createPath, data); err != nil {
		return nil, err
	}
	if err := t.render(destroyTemplateName, t.destroyPath, data); err != nil {
		return nil, err
	}
	return t, nil
}

func renderSetup(data SetupTemplateData) (*Terraform, error) {
	data.BucketPrefix = setupBucketPrefix
	t := &Terraform{
		createPath:  path.Join(rootPath, setupBucketPrefix, createDir),
		destroyPath: path.Join(rootPath, setupBucketPrefix, destroyDir),
	}
	if err := t.render(setupTemplateName, t.createPath, data); err != nil {
		return nil, err
	}
	if err := t.render(destroyTemplateName, t.destroyPath, data); err != nil {
		return nil, err
	}
	return t, nil
}

func extractModules() error {
	return extractEmbededDir(path.Join(rootPath, modulesDir), modulesDir)
}

func extractEmbededDir(path, name string) error {
	entries, err := fs.ReadDir(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	for _, e := range entries {
		local := path + "/" + e.Name()
		embeded := name + "/" + e.Name()
		if e.IsDir() {
			if err := extractEmbededDir(local, embeded); err != nil {
				return err
			}
		} else {
			content, err := fs.ReadFile(embeded)
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(local, content, os.ModePerm); err != nil {
				return err
			}
			//fmt.Printf("file written %s %d bytes\n", local, len(content))
		}
	}
	return nil
}

func (t *Terraform) render(name string, pth string, data interface{}) error {
	tpl, err := template.ParseFS(fs, path.Join(templatesDir, name))
	if err != nil {
		return err
	}
	funcs := template.FuncMap{"join": strings.Join}
	tpl = tpl.Funcs(funcs)

	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return err
	}
	if err := os.MkdirAll(pth, os.ModePerm); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(pth, mainTf), buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}
