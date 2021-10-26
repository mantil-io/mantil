package controller

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
)

const DestroyHTTPMethod = "destroy"

type StageArgs struct {
	Account    string
	Stage      string
	Force      bool
	DestroyAll bool
}

type Stage struct {
	store   *workspace.FileStore
	project *workspace.Project
	StageArgs
}

func NewStage(a StageArgs) (*Stage, error) {
	fs, err := NewStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &Stage{
		store:     fs,
		project:   fs.Project(),
		StageArgs: a,
	}, nil
}

func (s *Stage) New() error {
	if err := workspace.ValidateName(s.Stage); err != nil {
		return log.Wrap(err)
	}
	if s.Account == "" {
		accounts := s.store.Workspace().AccountNames()
		if len(accounts) > 1 {
			var err error
			s.Account, err = selectAccountForStage(accounts)
			if err != nil {
				return log.Wrap(err)
			}
		}
	}
	stage, err := s.store.Project().NewStage(s.Stage, s.Account)
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

func selectAccountForStage(accounts []string) (string, error) {
	prompt := promptui.Select{
		Label: "Select account for new stage",
		Items: accounts,
	}
	_, account, err := prompt.Run()
	if err != nil {
		return "", log.Wrap(err)
	}
	return account, nil
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

func (s *Stage) destroyStage(stage *workspace.Stage) error {
	ui.Info("Destroying stage %s in account %s", stage.Name, stage.Account().Name)
	if err := s.destroyRequest(stage); err != nil {
		return log.Wrap(err)
	}
	s.project.RemoveStage(stage.Name)
	return nil
}

func (s *Stage) destroyRequest(stage *workspace.Stage) error {
	account := stage.Account()
	req := &dto.DestroyRequest{
		Bucket:       account.Bucket,
		Region:       account.Region,
		ProjectName:  s.project.Name,
		StageName:    stage.Name,
		BucketPrefix: stage.BucketPrefix(),
		ResourceTags: stage.ResourceTags(),
	}
	backend, err := Backend(account)
	if err != nil {
		return log.Wrap(err)
	}
	if err := backend.Call(DestroyHTTPMethod, req, nil); err != nil {
		return log.Wrap(err)
	}
	return nil
}
