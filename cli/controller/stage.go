package controller

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/git"
	"github.com/mantil-io/mantil/github"
	"github.com/mantil-io/mantil/node/dto"
)

const DestroyHTTPMethod = "destroy"

const (
	CDPolicyManual   = "manual"
	CDPolicyOnCommit = "on-commit"
)

const (
	EnvIntegrationStage = "MANTIL_INTEGRATION_TOKEN"
	WorkflowFile        = "mantil-integration-workflow.yml"
)

//go:embed integration_workflow_template.yml
var integrationWorkflowTemplate string

type StageArgs struct {
	Node       string
	Stage      string
	Yes        bool
	DestroyAll bool
	CDPolicy   string
	CDToken    string
}

type Stage struct {
	store     *domain.FileStore
	project   *domain.Project
	ghClient  *github.Client
	gitClient *git.Client
	StageArgs
}

func NewStage(a StageArgs) (*Stage, error) {
	fs, project, err := newProjectStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	s := &Stage{
		store:     fs,
		project:   project,
		StageArgs: a,
	}
	if a.CDPolicy == CDPolicyOnCommit {
		if err := s.initIntegration(); err != nil {
			return nil, log.Wrap(err)
		}
	}
	return s, nil
}

func (s *Stage) initIntegration() error {
	if s.CDToken == "" {
		return log.Wrapf("cd token not specified")
	}
	var err error
	s.ghClient, err = github.New(s.CDToken)
	if err != nil {
		return log.Wrap(err)
	}
	s.gitClient, err = git.New(s.store.ProjectRoot(), s.CDToken)
	if err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (s *Stage) New() (bool, error) {
	if err := domain.ValidateName(s.Stage); err != nil {
		return false, log.Wrap(err)
	}

	// make sure there are nodes available for stage to be created on
	if len(s.store.Workspace().Nodes) == 0 {
		return false, log.Wrap(&domain.WorkspaceNoNodesError{})

	}

	// if node is specified make sure it exists or ask user to select one
	if s.Node != "" {
		node := s.store.Workspace().FindNode(s.Node)
		if node == nil {
			prompt := fmt.Sprintf("Node %s does not exist, please choose one of the available nodes for new stage", s.Node)
			s.Node = selectNodeForStage(prompt, s.store.Workspace().NodeNames())
			if s.Node == "" {
				return false, nil
			}
		}
	}

	// if node is not specified and there is more than one ask user to select one
	if s.Node == "" {
		nodes := s.store.Workspace().NodeNames()
		if len(nodes) > 1 {
			prompt := "There's more than one node available, please select one for new stage"
			s.Node = selectNodeForStage(prompt, nodes)
			if s.Node == "" {
				return false, nil
			}
		}
	}
	stage, err := s.chooseCreateStage()
	if err == promptui.ErrInterrupt {
		return false, nil
	}
	if err != nil {
		return false, log.Wrap(err)
	}
	if stage == nil {
		return false, nil
	}
	d, err := NewDeployWithStage(s.store, stage)
	if err != nil {
		return false, log.Wrap(err)
	}
	title := fmt.Sprintf("Creating stage %s on node %s", stage.Name, stage.NodeName)
	if err := d.DeployWithTitle(title); err != nil {
		return false, log.Wrap(err)
	}
	ui.Info("")
	ui.Title("Stage %s is ready!\n", stage.Name)
	ui.Info("Endpoint: %s", stage.RestEndpoint())

	if s.CDPolicy == CDPolicyOnCommit {
		if err := s.addGithubActionIntegration(stage); err != nil {
			return false, log.Wrap(err)
		}
	}
	return true, nil
}

func (s *Stage) chooseCreateStage() (*domain.Stage, error) {
	stageName := s.Stage
	var err error
	if stageName == "" {
		stageName, err = promptStageName()
		if err == promptui.ErrInterrupt {
			return nil, nil
		}
		if err != nil {
			return nil, log.Wrap(err)
		}

	}
	for {
		stage, err := s.project.NewStage(stageName, s.Node, s.store.ProjectRoot())
		var see *domain.StageExistsError
		if errors.As(err, &see) {
			ui.Info("Stage %s already exists", stageName)
			stageName, err = promptStageName()
			if err == promptui.ErrInterrupt {
				return nil, nil
			}
			if err != nil {
				return nil, log.Wrap(err)
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		return stage, nil
	}
}

func promptStageName() (string, error) {
	prompt := promptui.Prompt{
		Label: "Please specify stage name to continue",
	}
	stage, err := prompt.Run()
	return stage, err
}

func selectNodeForStage(text string, nodes []string) string {
	prompt := promptui.Select{
		Label: text,
		Items: nodes,
	}
	_, node, err := prompt.Run()
	if err != nil {
		return ""
	}
	return node
}

func (s *Stage) addGithubActionIntegration(stage *domain.Stage) error {
	if err := s.addGithubIntegrationToken(stage); err != nil {
		return log.Wrap(err)
	}
	if err := s.addGithubIntegrationWorkflow(stage.Name); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (s *Stage) addGithubIntegrationToken(stage *domain.Stage) error {
	ni, err := nodeInvoker(stage.Node())
	if err != nil {
		return log.Wrap(err)
	}
	req := &dto.AutomationJWTRequest{
		Project: s.project.Name,
		Stage:   stage.Name,
	}
	var rsp dto.AutomationJWTResponse
	if err := ni.Do("node/automationJWT", req, &rsp); err != nil {
		return log.Wrap(err)
	}
	od, err := s.gitClient.OriginData()
	if err != nil {
		return log.Wrap(err)
	}
	if err := s.ghClient.AddSecret(od.User, od.Repository, EnvIntegrationStage, rsp.Token); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Token was successfully added to the secrets of your repository")
	return nil
}

func (s *Stage) addGithubIntegrationWorkflow(stage string) error {
	if err := s.createWorkflowFile(stage); err != nil {
		return err
	}
	relPath := filepath.Join(".github", "workflows", WorkflowFile)
	if err := s.gitClient.Commit(relPath, "add mantil integration workflow"); err != nil {
		return err
	}
	ui.Info("Workflow file was successfully pushed to your repository")
	return nil
}

func (s *Stage) createWorkflowFile(stage string) error {
	branch, err := s.gitClient.Branch()
	if err != nil {
		return log.Wrap(err)
	}
	td := integrationWorkflowTemplateData{
		IntegrationStage: stage,
		EnvToken:         EnvIntegrationStage,
		Branch:           branch,
	}
	workflow, err := renderIntegrationWorkflowTemplate(integrationWorkflowTemplate, td)
	if err != nil {
		return log.Wrap(err)
	}
	destFolder := filepath.Join(s.store.ProjectRoot(), ".github", "workflows")
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

func (s *Stage) Destroy() error {
	if s.project.NumberOfStages() == 0 {
		return log.Wrap(&domain.ProjectNoStagesError{})
	}
	if s.DestroyAll {
		if !s.confirmDestroy() {
			return nil
		}
		if err := s.destroyAllStages(); err != nil {
			return log.Wrap(err)
		}
		ui.Info("")
		ui.Title("All stages were successfully destroyed!\n")
	} else {
		stage := s.chooseDestroyStage()
		if stage == nil {
			return nil
		}
		if !s.confirmDestroy() {
			return nil
		}
		if err := s.destroyStage(stage); err != nil {
			return log.Wrap(err)
		}
		ui.Info("")
		ui.Title("Stage %s was successfully destroyed!\n", stage.Name)
	}
	if err := s.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (s *Stage) chooseDestroyStage() *domain.Stage {
	stageName := s.Stage
	var err error
	for {
		stage := s.project.Stage(stageName)
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
		s.Stage = stage.Name
		return stage
	}
}

func (s *Stage) confirmDestroy() bool {
	if s.Yes {
		return true
	}
	if s.DestroyAll {
		ui.Title("? Do you really want to destroy all stages?\n")
	} else {
		ui.Title("? Do you really want to destroy stage %s?\n", s.Stage)
	}
	ui.Info("This action cannot be reversed.")
	confirmationPrompt := promptui.Prompt{
		Label: "To confirm, type 'yes'",
	}
	res, err := confirmationPrompt.Run()
	if err != nil {
		return false
	}
	res = strings.ToLower(res)
	if res != "yes" && res != "y" {
		return false
	}
	return true
}

func (s *Stage) destroyAllStages() error {
	var stages []*domain.Stage
	stages = append(stages, s.project.Stages...)
	for _, st := range stages {
		if err := s.destroyStage(st); err != nil {
			return log.Wrap(err)
		}
	}
	return nil
}

func (s *Stage) destroyStage(stage *domain.Stage) error {
	ui.Title("\nDestroying AWS infrastructure for stage %s\n", stage.Name)
	if err := s.destroyRequest(stage); err != nil {
		return log.Wrap(err)
	}
	s.project.RemoveStage(stage.Name)
	if err := s.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (s *Stage) destroyRequest(stage *domain.Stage) error {
	node := stage.Node()
	req := &dto.DestroyRequest{
		Bucket:                node.Bucket,
		Region:                node.Region,
		ProjectName:           s.project.Name,
		StageName:             stage.Name,
		BucketPrefix:          stage.StateBucketPrefix(),
		ResourceTags:          stage.ResourceTags(),
		CleanupBucketPrefixes: stage.BucketPrefixes(),
	}
	ni, err := nodeInvoker(node)
	if err != nil {
		return log.Wrap(err)
	}
	if err := ni.Do(DestroyHTTPMethod, req, nil); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (s *Stage) List() error {
	if len(s.project.Stages) == 0 {
		return log.Wrap(&domain.ProjectNoStagesError{})
	}
	var data [][]string
	for _, ps := range s.project.Stages {
		def := " "
		if ps.Default {
			def = "*"
		}
		data = append(data, []string{def, ps.Name, ps.NodeName, ps.RestEndpoint()})
	}
	ShowTable([]string{"default", "name", "node", "endpoint"}, data)
	return nil
}

func (s *Stage) Use() error {
	stage := s.project.Stage(s.Stage)
	if stage == nil {
		return log.Wrapf("stage %s not found", s.Stage)
	}
	s.project.SetDefaultStage(s.Stage)
	if err := s.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}
