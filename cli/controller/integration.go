package controller

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/git"
	"github.com/mantil-io/mantil/github"
	"github.com/mantil-io/mantil/node/dto"
)

const (
	EnvIntegrationStage = "MANTIL_INTEGRATION_TOKEN"
	EnvGithubToken      = "GITHUB_TOKEN"
	WorkflowFile        = "mantil-integration-workflow.yml"
)

//go:embed integration_workflow_template.yml
var integrationWorkflowTemplate string

type IntegrationArgs struct {
	Stage       string
	Repo        string
	GithubToken string
	GithubOrg   string
}

type Integration struct {
	store     *domain.FileStore
	project   *domain.Project
	ghClient  *github.Client
	gitClient *git.Client
	IntegrationArgs
}

// TODO
// - github org
// - automatic vs manual secret

func NewIntegration(a IntegrationArgs) (*Integration, error) {
	fs, project, err := newProjectStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	i := &Integration{
		store:           fs,
		project:         project,
		IntegrationArgs: a,
	}
	if a.Repo == "" {
		return i, nil
	}
	if i.GithubToken == "" {
		i.GithubToken = os.Getenv(EnvGithubToken)
	}
	if i.GithubToken == "" {
		return nil, fmt.Errorf("github token not specified")
	}
	i.ghClient, err = github.New(i.GithubToken, i.GithubOrg)
	if err != nil {
		return nil, log.Wrap(err)
	}
	i.gitClient, err = git.New(fs.ProjectRoot(), i.GithubToken)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (i *Integration) Add() error {
	stage := i.chooseStage()
	if stage == nil {
		return nil
	}
	i.Stage = stage.Name

	if err := i.addGithubIntegrationToken(stage); err != nil {
		return log.Wrap(err)
	}
	if err := i.addGithubIntegrationWorkflow(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (i *Integration) chooseStage() *domain.Stage {
	stageName := i.Stage
	var err error
	for {
		stage := i.project.Stage(stageName)
		if stage == nil {
			if stageName == "" {
				ui.Info("Stage name was not provided")
			} else {
				ui.Info("Stage %s doesn't exist", stageName)
			}
			stageName, err = promptStageName()
			if err != nil {
				return nil
			}
			continue
		}
		return stage
	}
}

func (i *Integration) addGithubIntegrationToken(stage *domain.Stage) error {
	ni, err := nodeInvoker(stage.Node())
	if err != nil {
		return log.Wrap(err)
	}
	req := &dto.AutomationJWTRequest{
		Project: i.project.Name,
		Stage:   stage.Name,
	}
	var rsp dto.AutomationJWTResponse
	if err := ni.Do("node/automationJWT", req, &rsp); err != nil {
		return log.Wrap(err)
	}
	if i.ghClient == nil {
		ui.Info("Your token is: %s", rsp.Token)
		ui.Info("Please add it to your repository secrets under the name %s.", EnvIntegrationStage)
		return nil
	}
	if err := i.ghClient.AddSecret(i.Repo, EnvIntegrationStage, rsp.Token); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Token was successfully added to the secrets of your repository.")
	return nil
}

func (i *Integration) addGithubIntegrationWorkflow() error {
	if err := i.createWorkflowFile(); err != nil {
		return err
	}
	if i.gitClient == nil {
		return nil
	}
	relPath := filepath.Join(".github", "workflows", WorkflowFile)
	if err := i.gitClient.Commit(relPath, "add mantil integration workflow"); err != nil {
		return err
	}
	ui.Info("Workflow file was successfully pushed to your repository.")
	return nil
}

func (i *Integration) createWorkflowFile() error {
	branch, err := i.gitClient.Branch()
	if err != nil {
		return err
	}
	td := integrationWorkflowTemplateData{
		IntegrationStage: i.Stage,
		EnvToken:         EnvIntegrationStage,
		Branch:           branch,
	}
	workflow, err := renderIntegrationWorkflowTemplate(integrationWorkflowTemplate, td)
	if err != nil {
		return log.Wrap(err)
	}
	destFolder := filepath.Join(i.store.ProjectRoot(), ".github", "workflows")
	if err := os.MkdirAll(destFolder, os.ModePerm); err != nil {
		return log.Wrap(err)
	}
	destFile := filepath.Join(destFolder, WorkflowFile)
	if err := ioutil.WriteFile(destFile, workflow, 0644); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func renderIntegrationWorkflowTemplate(content string, data integrationWorkflowTemplateData) ([]byte, error) {
	tpl, err := template.New("").Delims("[[", "]]").Parse(content)
	if err != nil {
		return nil, log.Wrap(err)
	}
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return nil, log.Wrap(err)
	}
	return buf.Bytes(), nil
}

type integrationWorkflowTemplateData struct {
	IntegrationStage string
	EnvToken         string
	Branch           string
}
