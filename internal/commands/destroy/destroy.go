package destroy

import (
	"fmt"
	"os"

	"github.com/mantil-io/mantil-cli/internal/commands"
	"github.com/mantil-io/mantil-cli/internal/github"
	"github.com/mantil-io/mantil-cli/internal/log"
	"github.com/mantil-io/mantil-cli/internal/mantil"
)

type DestroyCmd struct {
	project   *mantil.Project
	githubOrg string
	path      string
	token     string
}

func New(project *mantil.Project, githubOrg, path, token string) (*DestroyCmd, error) {
	return &DestroyCmd{
		project:   project,
		githubOrg: githubOrg,
		path:      path,
		token:     token,
	}, nil
}

func (d *DestroyCmd) Destroy(deleteRepo bool) error {
	log.Info("Destroying infrastructure...")
	err := d.destroyRequest()
	if err != nil {
		return fmt.Errorf("could not destroy infrastructure - %v", err)
	}
	if deleteRepo {
		if err := d.deleteRepo(); err != nil {
			return err
		}
	}
	log.Notice("destroy successfully finished")
	return nil
}

func (d *DestroyCmd) deleteRepo() error {
	log.Info("Deleting local files...")
	os.RemoveAll(d.path)
	log.Info("Deleting github repository...")
	ghClient, err := github.NewClient(d.githubOrg)
	if err != nil {
		return fmt.Errorf("could not initialize github client - %v", err)
	}
	name := d.project.Name
	err = ghClient.DeleteRepo(name)
	if err != nil {
		return fmt.Errorf("could not delete repo %s - %v", name, err)
	}
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
