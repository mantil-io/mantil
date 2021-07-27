package destroy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
)

type DestroyCmd struct {
	aws       *aws.AWS
	project   *mantil.Project
	githubOrg string
	path      string
	token     string
}

func New(project *mantil.Project, githubOrg string, path string) (*DestroyCmd, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	token, err := mantil.ReadToken(project.Name)
	if err != nil {
		return nil, err
	}
	return &DestroyCmd{
		aws:       awsClient,
		project:   project,
		githubOrg: githubOrg,
		path:      path,
		token:     token,
	}, nil
}

func (d *DestroyCmd) Destroy() error {
	log.Println("Destroying infrastructure...")
	err := d.destroyRequest()
	if err != nil {
		return fmt.Errorf("could not destroy infrastructure - %v", err)
	}
	log.Println("Deleting local files...")
	os.RemoveAll(d.path)
	log.Println("Deleting github repository...")
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
	url := "https://try.mantil.team/mantil-backend/destroy"
	r := &req{
		ProjectName: d.project.Name,
		Token:       d.token,
	}
	buf, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	return nil
}
