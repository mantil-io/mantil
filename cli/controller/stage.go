package controller

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/dto"
	"github.com/olekukonko/tablewriter"
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

func (s *Stage) New() error {
	if err := domain.ValidateName(s.Stage); err != nil {
		return log.Wrap(err)
	}

	// make sure there are nodes available for stage to be created on
	if len(s.store.Workspace().Nodes) == 0 {
		return log.Wrapf("No nodes currently exist, please first create one with 'mantil aws install'")

	}

	// if node is specified make sure it exists or ask user to select one
	if s.Node != "" {
		node := s.store.Workspace().FindNode(s.Node)
		if node == nil {
			var err error
			prompt := fmt.Sprintf("Node %s does not exist, please choose one of the available nodes for new stage", s.Node)
			s.Node, err = selectNodeForStage(prompt, s.store.Workspace().NodeNames())
			if err != nil {
				return log.Wrap(err)
			}
		} else {
			ui.Info("Using node %s for new stage", s.Node)
		}
	}

	// if node is not specified and there is more than one ask user to select one
	if s.Node == "" {
		nodes := s.store.Workspace().NodeNames()
		if len(nodes) > 1 {
			var err error
			prompt := "There's more than one node available, please select one for new stage"
			s.Node, err = selectNodeForStage(prompt, nodes)
			if err != nil {
				return log.Wrap(err)
			}
		}
	}
	stage, err := s.project.NewStage(s.Stage, s.Node, s.store.ProjectRoot())
	if err != nil {
		return log.Wrap(err)
	}
	// node wasn't specified or chosen above, default one is used
	if s.Node == "" {
		ui.Info("Using node %s as default for new stage", stage.Node().Name)
	}
	ui.Info("")
	ui.Title("Creating stage %s and deploying project %s\n", stage.Name, stage.Project().Name)
	d, err := NewDeployWithStage(s.store, stage)
	if err != nil {
		return log.Wrap(err)
	}
	if err := d.Deploy(); err != nil {
		return log.Wrap(err)
	}
	ui.Info("")
	ui.Title("Stage %s is ready!\n", stage.Name)
	ui.Info("Endpoint: %s", stage.Endpoints.Rest)
	return nil
}

func selectNodeForStage(text string, nodes []string) (string, error) {
	prompt := promptui.Select{
		Label: text,
		Items: nodes,
	}
	_, node, err := prompt.Run()
	if err != nil {
		return "", log.Wrap(err)
	}
	return node, nil
}

func (s *Stage) Destroy() error {
	if !s.DestroyAll && s.Stage == "" {
		return log.Wrapf("No stage specified")
	}
	if s.DestroyAll {
		if !s.confirmDestroy() {
			return nil
		}
		for _, stage := range s.project.Stages {
			if err := s.destroyStage(stage); err != nil {
				return log.Wrap(err)
			}
		}
		ui.Info("")
		ui.Title("All stages were successfully destroyed!\n")
	} else {
		stage := s.project.Stage(s.Stage)
		if stage == nil {
			return log.Wrapf("Stage %s not found", s.Stage)
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

func (s *Stage) destroyStage(stage *domain.Stage) error {
	ui.Title("\nDestroying AWS infrastructure\n")
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
		Bucket:       node.Bucket,
		Region:       node.Region,
		ProjectName:  s.project.Name,
		StageName:    stage.Name,
		BucketPrefix: stage.StateBucketPrefix(),
		ResourceTags: stage.ResourceTags(),
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
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"default", "name", "node", "endpoint"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	for _, ps := range s.project.Stages {
		def := " "
		if ps.Default {
			def = "*"
		}
		table.Append([]string{def, ps.Name, ps.NodeName, ps.Endpoints.Rest})
	}
	table.Render()
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
