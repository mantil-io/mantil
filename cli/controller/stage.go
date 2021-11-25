package controller

import (
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/dto"
)

const DestroyHTTPMethod = "destroy"

type StageArgs struct {
	Node       string
	Stage      string
	Yes        bool
	DestroyAll bool
}

type Stage struct {
	store   *domain.FileStore
	project *domain.Project
	StageArgs
}

func NewStage(a StageArgs) (*Stage, error) {
	fs, project, err := newProjectStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &Stage{
		store:     fs,
		project:   project,
		StageArgs: a,
	}, nil
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
	ui.Info("Endpoint: %s", stage.Endpoints.Rest)
	return true, nil
}

func (s *Stage) chooseCreateStage() (*domain.Stage, error) {
	stageName := s.Stage
	for {
		stage, err := s.project.NewStage(stageName, s.Node, s.store.ProjectRoot())
		var see *domain.StageExistsError
		if errors.As(err, &see) {
			ui.Info("Stage %s already exists", stageName)
			stageName, err = promptStageName()
			if err != nil {
				return nil, nil
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
		Label: "Please specify a new stage name to continue",
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
		data = append(data, []string{def, ps.Name, ps.NodeName, ps.Endpoints.Rest})
	}
	ShowTable([]string{"default", "name", "node", "endpoint"}, data)
	return nil
}

func (s *Stage) Use() error {
	stage := s.project.Stage(s.Stage)
	if stage == nil {
		return log.Wrapf("Stage %s not found", s.Stage)
	}
	s.project.SetDefaultStage(s.Stage)
	if err := s.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}
