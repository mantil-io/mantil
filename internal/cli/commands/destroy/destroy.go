package destroy

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/cli/commands"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/config"
	"github.com/mantil-io/mantil/internal/git"
)

type DestroyCmd struct {
	project *config.Project
	path    string
	stage   string
}

func New(project *config.Project, path, stage string) (*DestroyCmd, error) {
	return &DestroyCmd{
		project: project,
		path:    path,
		stage:   stage,
	}, nil
}

func (d *DestroyCmd) Destroy(deleteRepo bool) error {
	log.Info("Destroying infrastructure...")
	err := d.destroyRequest()
	if err != nil {
		return fmt.Errorf("could not destroy infrastructure - %v", err)
	}
	if deleteRepo {
		log.Info("Deleting local repository...")
		if err := git.DeleteRepo(d.path); err != nil {
			return err
		}
	}
	if d.stage != "" {
		d.project.RemoveStage(d.stage)
		config.SaveProject(d.project, d.path)
	}
	log.Notice("destroy successfully finished")
	return nil
}

func (d *DestroyCmd) destroyRequest() error {
	type req struct {
		ProjectName string
		Stage       string
	}
	r := &req{
		ProjectName: d.project.Name,
		Stage:       d.stage,
	}
	if err := commands.BackendRequest("destroy", r, nil, true); err != nil {
		return err
	}
	return nil
}
