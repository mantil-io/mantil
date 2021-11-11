package controller

import (
	"fmt"
	"os"

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
	Force      bool
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
	if s.Node == "" {
		nodes := s.store.Workspace().NodeNames()
		if len(nodes) > 1 {
			var err error
			s.Node, err = selectNodeForStage(nodes)
			if err != nil {
				return log.Wrap(err)
			}
		}
	}
	stage, err := s.store.Project().NewStage(s.Stage, s.Node)
	if err != nil {
		return log.Wrap(err)
	}
	d, err := NewDeployWithStage(s.store, stage)
	if err != nil {
		return log.Wrap(err)
	}
	if err := d.Deploy(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func selectNodeForStage(nodes []string) (string, error) {
	prompt := promptui.Select{
		Label: "Select node for new stage",
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
	if !s.Force {
		if err := s.confirmDestroy(); err != nil {
			return log.Wrap(err)
		}
	}
	if s.DestroyAll {
		for _, stage := range s.project.Stages {
			if err := s.destroyStage(stage); err != nil {
				return log.Wrap(err)
			}
		}
	} else {
		stage := s.project.Stage(s.Stage)
		if stage == nil {
			return log.Wrap(fmt.Errorf("Stage %s not found", s.Stage))
		}
		if err := s.destroyStage(stage); err != nil {
			return log.Wrap(err)
		}
	}
	if err := s.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (s *Stage) confirmDestroy() error {
	var label string
	if s.DestroyAll {
		label = "To confirm deletion of all stages, please enter the project name"
	} else {
		label = fmt.Sprintf("To confirm deletion of stage %s, please enter the project name", s.Stage)
	}
	confirmationPrompt := promptui.Prompt{
		Label: label,
	}
	projectName, err := confirmationPrompt.Run()
	if err != nil {
		return log.Wrap(err)
	}
	if s.project.Name != projectName {
		return log.Wrap(err)
	}
	return nil
}

func (s *Stage) destroyStage(stage *domain.Stage) error {
	ui.Info("Destroying stage %s in node %s", stage.Name, stage.Node().Name)
	if err := s.destroyRequest(stage); err != nil {
		return log.Wrap(err)
	}
	s.project.RemoveStage(stage.Name)
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

func Nodes() error {
	fs, err := newStore()
	if err != nil {
		return log.Wrap(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"name", "AWS Account", "AWS Region", "ID"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	for _, n := range fs.Workspace().Nodes {
		table.Append([]string{n.Name, n.ID, n.Region, n.UID})
	}
	table.Render()
	return nil
}
