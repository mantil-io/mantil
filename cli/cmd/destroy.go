package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/commands"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/git"
)

type destroyCmd struct {
	ctx        *commands.ProjectContext
	deleteRepo bool
}

func (c *destroyCmd) run() error {
	log.Info("Destroying infrastructure...")
	err := c.destroyRequest()
	if err != nil {
		return fmt.Errorf("could not destroy infrastructure - %v", err)
	}
	if c.deleteRepo {
		log.Info("Deleting local repository...")
		if err := git.DeleteRepo(c.ctx.Path); err != nil {
			return err
		}
	}
	c.ctx.Project.RemoveStage(c.ctx.Stage.Name)
	config.SaveProject(c.ctx.Project, c.ctx.Path)
	log.Notice("Destroy successfully finished")
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
