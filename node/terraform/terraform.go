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
	"github.com/mantil-io/mantil/kit/shell"
	"github.com/mantil-io/mantil/node/dto"
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
	pluginsDir          = "plugins"
)

type Terraform struct {
	path           string
	createPath     string
	destroyPath    string
	pluginsPath    string
	createContent  []byte
	destroyContent []byte
	parser         *Parser
}

func New(createPath, destroyPath string) (*Terraform, error) {
	pluginsPath, err := createPluginsDir()
	if err != nil {
		return nil, err
	}
	return &Terraform{
		createPath:  createPath,
		destroyPath: destroyPath,
		pluginsPath: pluginsPath,
		parser:      NewLogParser(),
	}, nil
}

type SetupTemplateData struct {
	Bucket          string
	BucketPrefix    string
	FunctionsBucket string
	FunctionsPath   string
	Region          string
	ResourceSuffix  string
	NamingTemplate  string
	AuthEnv         map[string]string
	ResourceTags    map[string]string
}

// Prepare setup templates
func Setup(data SetupTemplateData) (*Terraform, error) {
	if err := extractModules(); err != nil {
		return nil, err
	}
	return renderSetup(data)
}

// Prepare project templates
func Project(data dto.StageTemplate) (*Terraform, error) {
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
		return t.shellExecDefault(args)
	}
	return nil
}

func (t *Terraform) apply(destroy bool) error {
	args := []string{"terraform", "apply", "-no-color", "-input=false", "-compact-warnings", "-auto-approve"}
	if destroy {
		args = append(args, "-destroy")
	}
	opt := t.shellExecOpts(logPrefix, args)
	opt.ErrorsMap = map[string]error{
		"ConflictException: Unable to complete operation due to concurrent modification. Please try again later.": conflictException,
	}
	return t.shellExec(opt)
}

func (t *Terraform) output() error {
	args := []string{"terraform", "output", "-no-color"}
	return t.shellExec(t.shellExecOpts(outputLogPrefix, args))
}

/////////////// shell exec

var conflictException = fmt.Errorf("ConflictException")

func (t *Terraform) shellExecOpts(logPrefix string, args []string) shell.ExecOptions {
	opt := shell.ExecOptions{
		Args:         args,
		WorkDir:      t.path,
		ShowShellCmd: true,
		Env: []string{
			"TF_IN_AUTOMATION=true",
			"TF_PLUGIN_CACHE_DIR=" + t.pluginsPath,
		},
		Logger: func(format string, v ...interface{}) {
			format = logPrefix + format
			line := fmt.Sprintf(format, v...)
			t.parser.Parse(line)
			stdlog.Print(line)
		},
	}
	if p := aws.TestProfile(); p != "" {
		opt.Env = []string{"AWS_PROFILE=" + p}
	}
	return opt
}

func (t *Terraform) shellExecDefault(args []string) error {
	return t.shellExec(t.shellExecOpts(logPrefix, args))
}

func (t *Terraform) shellExec(opts shell.ExecOptions) error {
	err := shell.Exec(opts)
	if perr := t.parser.Error(); perr != nil {
		return perr
	}
	return err
}

/////////////// rendering templates

func renderProject(data dto.StageTemplate) (*Terraform, error) {
	stageDir := fmt.Sprintf("%s-%s", data.Project, data.Stage)
	t, err := New(
		path.Join(rootPath, stageDir, createDir),
		path.Join(rootPath, stageDir, destroyDir),
	)
	if err != nil {
		return nil, err
	}
	if t.createContent, err = t.render(projectTemplateName, t.createPath, data); err != nil {
		return nil, err
	}
	if t.destroyContent, err = t.render(destroyTemplateName, t.destroyPath, data); err != nil {
		return nil, err
	}
	return t, nil
}

func renderSetup(data SetupTemplateData) (*Terraform, error) {
	data.BucketPrefix = setupBucketPrefix
	t, err := New(
		path.Join(rootPath, setupBucketPrefix, createDir),
		path.Join(rootPath, setupBucketPrefix, destroyDir),
	)
	if err != nil {
		return nil, err
	}
	if t.createContent, err = t.render(setupTemplateName, t.createPath, data); err != nil {
		return nil, err
	}
	if t.destroyContent, err = t.render(destroyTemplateName, t.destroyPath, data); err != nil {
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

func (t *Terraform) render(name string, pth string, data interface{}) ([]byte, error) {
	tpl, err := template.ParseFS(fs, path.Join(templatesDir, name))
	if err != nil {
		return nil, err
	}
	funcs := template.FuncMap{"join": strings.Join}
	tpl = tpl.Funcs(funcs)

	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(pth, os.ModePerm); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(path.Join(pth, mainTf), buf.Bytes(), 0644); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (t *Terraform) Output(key string) (string, error) {
	val, ok := t.parser.Outputs[key]
	if !ok {
		return "", fmt.Errorf("output variable %s not found", key)
	}
	return val, nil
}

func createPluginsDir() (string, error) {
	path := path.Join(rootPath, pluginsDir)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}
	return path, nil
}
