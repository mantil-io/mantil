package initialize

import (
	"fmt"
	"path/filepath"

	"github.com/mantil-io/mantil-cli/internal/commands"
	"github.com/mantil-io/mantil-cli/internal/github"
	"github.com/mantil-io/mantil-cli/internal/log"
	"github.com/mantil-io/mantil-cli/internal/mantil"
)

type InitCmd struct {
	name      string
	githubOrg string
}

func New(name, githubOrg string) (*InitCmd, error) {
	return &InitCmd{
		name:      name,
		githubOrg: githubOrg,
	}, nil
}

func (i *InitCmd) InitProject() error {
	token, err := i.initRequest(i.name)
	if err != nil || token == "" {
		return fmt.Errorf("could not initialize project - %v", err)
	}
	log.Info("Creating repo from template...")
	githubClient, err := github.NewClient(i.githubOrg)
	if err != nil {
		return fmt.Errorf("could not initialize github client - %v", err)
	}
	templateRepo := "https://github.com/mantil-io/go-mantil-template"
	project, err := mantil.NewProject(i.name)
	if err != nil {
		return fmt.Errorf("could not create project %s - %v", i.name, err)
	}
	lc := project.LocalConfig(i.githubOrg)
	repoURL, err := githubClient.CreateRepoFromTemplate(templateRepo, i.name, i.name, lc)
	if err != nil {
		return fmt.Errorf("could not create repo %s from template - %v", i.name, err)
	}
	if err := githubClient.AddSecrets(i.name, token); err != nil {
		return fmt.Errorf("could not add mantil token to repo - %v", err)
	}
	if err := mantil.SaveToken(i.name, token); err != nil {
		return fmt.Errorf("could not save token to ~/.mantil directory - %v", err)
	}
	projectPath, _ := filepath.Abs(i.name)
	log.Notice("Done!\nProject initialized at %s\nGithub repo URL: %s", projectPath, repoURL)
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
