package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/commands"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/git"
)

type destroyCmd struct {
	project    *config.Project
	stageName  string
	repoPath   string
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
		if err := git.DeleteRepo(c.repoPath); err != nil {
			return err
		}
	}
	c.project.RemoveStage(c.stageName)
	config.SaveProject(c.project, c.repoPath)
	log.Notice("Destroy successfully finished")
	return nil
}

func (c *destroyCmd) destroyRequest() error {
	type req struct {
		ProjectName string
		StageName   string
	}
	r := &req{
		ProjectName: c.project.Name,
		StageName:   c.stageName,
	}
	if err := commands.BackendRequest("destroy", r, nil, true); err != nil {
		return err
	}
	return nil
}
