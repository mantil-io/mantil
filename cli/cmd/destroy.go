package cmd

import (
	"fmt"
	"github.com/mantil-io/mantil/cli/cmd/project"

	"github.com/mantil-io/mantil/cli/log"
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
		if err := c.destroyStage(c.ctx.Stage); err != nil {
			return err
		}
	}
	if c.deleteRepo {
		log.Info("Deleting local repository...")
		if err := git.DeleteRepo(c.ctx.Path); err != nil {
			return err
		}
	}
	workspace.SaveProject(c.ctx.Project, c.ctx.Path)
	log.Notice("Destroy successfully finished")
	return nil
}

func (c *destroyCmd) destroyStage(stage *workspace.Stage) error {
	if err := c.ctx.SetStage(stage); err != nil {
		return err
	}
	log.Info("Destroying stage %s in account %s", c.ctx.Stage.Name, c.ctx.Account.Name)
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
