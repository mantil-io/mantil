package destroy

import (
	"fmt"
	"os"

	"github.com/atoz-technology/mantil-cli/internal/commands"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/internal/log"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
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

func (d *DestroyCmd) Destroy() error {
	log.Info("Destroying infrastructure...")
	err := d.destroyRequest()
	if err != nil {
		return fmt.Errorf("could not destroy infrastructure - %v", err)
	}
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
