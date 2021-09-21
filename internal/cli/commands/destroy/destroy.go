package destroy

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/cli/commands"
	"github.com/mantil-io/mantil/internal/cli/git"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/mantil"
)

type DestroyCmd struct {
	project *mantil.Project
	path    string
	token   string
}

func New(project *mantil.Project, path, token string) (*DestroyCmd, error) {
	return &DestroyCmd{
		project: project,
		path:    path,
		token:   token,
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
	log.Notice("destroy successfully finished")
	return nil
}

func (d *DestroyCmd) destroyRequest() error {
	type req struct {
		ProjectName string
		Token       string
	}
	r := &req{
		ProjectName: d.project.Name,
		Token:       d.token,
	}
	if err := commands.BackendRequest("destroy", r, nil); err != nil {
		return err
	}
	return nil
}
