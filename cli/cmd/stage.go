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

type stageFlags struct {
	account    string
	stage      string
	force      bool
	destroyAll bool
}

type stageCmd struct {
	ctx        *project.Context
	stage      string
	account    string
	force      bool
	destroyAll bool
}

func newStage(f *stageFlags) (*stageCmd, error) {
	ctx, err := project.NewContext()
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &stageCmd{
		ctx:        ctx,
		stage:      f.stage,
		account:    f.account,
		force:      f.force,
		destroyAll: f.destroyAll,
	}, nil
}

func (c *stageCmd) new() error {
	if c.stage == "" {
		c.stage = workspace.DefaultStageName
	}
	if s := c.ctx.Project.Stage(c.stage); s != nil {
		return log.WithUserMessage(nil, fmt.Sprintf("Stage %s already exists.", c.stage))
	}
	ui.Info("Creating stage %s...", c.stage)
	stage, err := c.createStage(c.account)
	if err != nil {
		return log.Wrap(err)
	}
	if err = c.ctx.SetStage(stage); err != nil {
		return log.Wrap(err)
	}
	d, err := deploy.NewFromContext(c.ctx)
	if err != nil {
		return log.Wrap(err)
	}
	_, err = d.Deploy()
	if err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *stageCmd) createStage(accountName string) (*workspace.Stage, error) {
	if len(c.ctx.Workspace.Accounts) == 0 {
		return nil, log.WithUserMessage(nil, "No accounts found in workspace. Please add an account with `mantil install`.")
	}
	if accountName == "" {
		var err error
		if len(c.ctx.Workspace.Accounts) > 1 {
			accountName, err = c.selectAccount()
			if err != nil {
				return nil, log.Wrap(err)
			}
		} else {
			accountName = c.ctx.Workspace.Accounts[0].Name
		}
	}
	stage := &workspace.Stage{
		Name:    c.stage,
		Account: accountName,
	}
	if len(c.ctx.Project.Stages) == 0 {
		stage.Default = true
	}
	return stage, nil
}

func (s *stageCmd) selectAccount() (string, error) {
	var accounts []string
	for _, a := range s.ctx.Workspace.Accounts {
		accounts = append(accounts, a.Name)
	}
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
		for _, s := range c.ctx.Project.Stages {
			if err := c.destroyStage(s); err != nil {
				return err
			}
		}
	} else {
		s := c.ctx.Project.Stage(c.stage)
		if s == nil {
			return log.Wrap(fmt.Errorf("stage %s not found", c.stage))
		}
		if err := c.destroyStage(s); err != nil {
			return log.Wrap(err)
		}
	}
	c.ctx.Project.SetDefaultStage()
	workspace.SaveProject(c.ctx.Project, c.ctx.Path)
	ui.Notice("Destroy successfully finished")
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
	if c.ctx.Project.Name != projectName {
		return log.Wrap(err)
	}
	return nil
}

func (c *stageCmd) destroyStage(stage *workspace.Stage) error {
	if err := c.ctx.SetStage(stage); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Destroying stage %s in account %s", c.ctx.Stage.Name, c.ctx.Account.Name)
	if err := c.destroyRequest(); err != nil {
		return log.Wrap(fmt.Errorf("could not destroy stage %s - %v", c.ctx.Stage.Name, err))
	}
	c.ctx.Project.RemoveStage(c.ctx.Stage.Name)
	return nil
}

func (c *stageCmd) destroyRequest() error {
	req := &dto.DestroyRequest{
		ProjectName: c.ctx.Project.Name,
		StageName:   c.ctx.Stage.Name,
	}
	if err := c.ctx.RuntimeRequest("destroy", req, nil, true); err != nil {
		return log.Wrap(err)
	}
	return nil
}
