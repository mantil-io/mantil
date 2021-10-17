package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
)

const DestroyHTTPMethod = "destroy"

type stageArgs struct {
	account    string
	stage      string
	force      bool
	destroyAll bool
}

type stageCmd struct {
	store   *workspace.FileStore
	project *workspace.Project
	stageArgs
}

func newStage(a stageArgs) (*stageCmd, error) {
	fs, err := project.NewStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &stageCmd{
		store:     fs,
		project:   fs.Project(),
		stageArgs: a,
	}, nil
}

func (c *stageCmd) new() error {
	if c.account == "" {
		accounts := c.store.Workspace().AccountNames()
		if len(accounts) > 1 {
			var err error
			c.account, err = c.selectAccount(accounts)
			if err != nil {
				return log.Wrap(err)
			}
		}
	}

	stage, err := c.project.NewStage(c.stage, c.account)
	if err != nil {
		if err == workspace.ErrStageExists {
			return log.WithUserMessage(err, "Stage %s already exists.", c.stage)
		}
		if err == workspace.ErrAccountNotFound {
			return log.WithUserMessage(err, "Account %s not found.", c.account)
		}
		return log.Wrap(err)
	}

	d, err := deploy.NewWithStage(c.store, stage)
	if err != nil {
		return log.Wrap(err)
	}
	if err := d.Deploy(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (s *stageCmd) selectAccount(accounts []string) (string, error) {
	prompt := promptui.Select{
		Label: "Select an account",
		Items: accounts,
	}
	_, account, err := prompt.Run()
	if err != nil {
		return "", log.Wrap(err)
	}
	return account, nil
}

func (c *stageCmd) destroy() error {
	if !c.destroyAll && c.stage == "" {
		return log.WithUserMessage(nil, "No stage specified")
	}
	if !c.force {
		if err := c.confirmDestroy(); err != nil {
			return log.Wrap(err)
		}
	}
	if c.destroyAll {
		for _, s := range c.project.Stages {
			if err := c.destroyStage(s); err != nil {
				return err
			}
		}
	} else {
		s := c.project.Stage(c.stage)
		if s == nil {
			return log.Wrap(fmt.Errorf("Stage %s not found", c.stage))
		}
		if err := c.destroyStage(s); err != nil {
			return log.Wrap(err)
		}
	}
	c.project.SetDefaultStage()
	c.store.Store()
	//ui.Notice("Destroy successfully finished")
	return nil
}

func (c *stageCmd) confirmDestroy() error {
	var label string
	if c.destroyAll {
		label = "To confirm deletion of all stages, please enter the project name"
	} else {
		label = fmt.Sprintf("To confirm deletion of stage %s, please enter the project name", c.stage)
	}
	confirmationPrompt := promptui.Prompt{
		Label: label,
	}
	projectName, err := confirmationPrompt.Run()
	if err != nil {
		return log.Wrap(err)
	}
	if c.project.Name != projectName {
		return log.Wrap(err)
	}
	return nil
}

func (c *stageCmd) destroyStage(stage *workspace.Stage) error {
	ui.Info("Destroying stage %s in account %s", stage.Name, stage.Account().Name)
	if err := c.destroyRequest(stage); err != nil {
		return log.Wrap(err)
	}
	c.project.RemoveStage(stage.Name)
	return nil
}

func (c *stageCmd) destroyRequest(stage *workspace.Stage) error {
	account := stage.Account()
	req := &dto.DestroyRequest{
		Bucket:       account.Bucket,
		Region:       account.Region,
		ProjectName:  c.project.Name,
		StageName:    stage.Name,
		BucketPrefix: stage.BucketPrefix(),
	}
	backend, err := project.Backend(account)
	if err != nil {
		return log.Wrap(err)
	}
	if err := backend.Call(DestroyHTTPMethod, req, nil); err != nil {
		return log.Wrap(err)
	}
	return nil
}
