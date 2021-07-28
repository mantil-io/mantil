package initialize

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/commands"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
)

type InitCmd struct {
	aws       *aws.AWS
	name      string
	githubOrg string
}

func New(name, githubOrg string) (*InitCmd, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, fmt.Errorf("could not initialize aws - %v", err)
	}

	return &InitCmd{
		aws:       awsClient,
		name:      name,
		githubOrg: githubOrg,
	}, nil
}

func (i *InitCmd) InitProject() error {
	token, err := i.initRequest(i.name)
	if err != nil || token == "" {
		return fmt.Errorf("could not initialize project - %v", err)
	}
	log.Println("Creating repo from template...")
	githubClient, err := github.NewClient(i.githubOrg)
	if err != nil {
		return fmt.Errorf("could not initialize github client - %v", err)
	}
	templateRepo := "https://github.com/atoz-technology/go-mantil-template"
	project, err := mantil.NewProject(i.name)
	if err != nil {
		return fmt.Errorf("could not create project %s - %v", i.name, err)
	}
	lc := project.LocalConfig(i.githubOrg)
	repoURL, err := githubClient.CreateRepoFromTemplate(templateRepo, i.name, i.name, lc)
	if err != nil {
		return fmt.Errorf("could not create repo %s from template - %v", i.name, err)
	}
	if err := githubClient.AddSecrets(i.name, i.aws, token); err != nil {
		return fmt.Errorf("could not add mantil token to repo - %v", err)
	}
	if err := mantil.SaveToken(i.name, token); err != nil {
		return fmt.Errorf("could not save token to ~/.mantil directory - %v", err)
	}
	projectPath, _ := filepath.Abs(i.name)
	log.Printf("Done!\nProject initialized at %s\nGithub repo URL: %s", projectPath, repoURL)
	return nil
}

func (i *InitCmd) initRequest(projectName string) (string, error) {
	type initReq struct {
		ProjectName string
	}
	ireq := &initReq{
		ProjectName: projectName,
	}
	type initResp struct {
		Token string
	}
	iresp := &initResp{}
	if err := commands.BackendRequest("init", ireq, iresp); err != nil {
		return "", err
	}
	return iresp.Token, nil
}
