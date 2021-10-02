package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/cmd/project"

	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/git"
	"github.com/mantil-io/mantil/workspace"
)

type destroyCmd struct {
	stageName  string
	ctx        *project.Context
	deleteRepo bool
}

func (c *destroyCmd) run() error {
	if c.stageName == "" {
		for _, s := range c.ctx.Project.Stages {
			if err := c.destroyStage(s); err != nil {
				return err
			}
		}
	} else {
		s := c.ctx.Project.Stage(c.stageName)
		if s == nil {
			return fmt.Errorf("stage %s not found", c.stageName)
		}
		if err := c.destroyStage(s); err != nil {
			return err
		}
	}
	if c.deleteRepo {
		ui.Info("Deleting local repository...")
		if err := git.DeleteRepo(c.ctx.Path); err != nil {
			return err
		}
	}
	c.ctx.Project.SetDefaultStage()
	workspace.SaveProject(c.ctx.Project, c.ctx.Path)
	ui.Notice("Destroy successfully finished")
	return nil
}

func (c *destroyCmd) destroyStage(stage *workspace.Stage) error {
	if err := c.ctx.SetStage(stage); err != nil {
		return err
	}
	ui.Info("Destroying stage %s in account %s", c.ctx.Stage.Name, c.ctx.Account.Name)
	if err := c.destroyRequest(); err != nil {
		return fmt.Errorf("could not destroy stage %s - %v", c.ctx.Stage.Name, err)
	}
	c.ctx.Project.RemoveStage(c.ctx.Stage.Name)
	return nil
}

func (c *destroyCmd) destroyRequest() error {
	type req struct {
		ProjectName string
		StageName   string
	}
	r := &req{
		ProjectName: c.ctx.Project.Name,
		StageName:   c.ctx.Stage.Name,
	}
	if err := c.ctx.RuntimeRequest("destroy", r, nil, true); err != nil {
		return err
	}
	return nil
}
