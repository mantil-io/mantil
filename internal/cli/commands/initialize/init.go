package initialize

import (
	"fmt"
	"path/filepath"

	"github.com/mantil-io/mantil/internal/cli/commands"
	"github.com/mantil-io/mantil/internal/cli/git"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/mantil"
)

type InitCmd struct {
	name      string
	noRepo    bool
	githubOrg string
	template  string
}

func New(name, githubOrg string, noRepo bool, template string) (*InitCmd, error) {
	return &InitCmd{
		name:      name,
		noRepo:    noRepo,
		githubOrg: githubOrg,
		template:  template,
	}, nil
}

func (i *InitCmd) InitProject() error {
	templateRepo := i.templateRepo()
	if templateRepo == "" {
		return fmt.Errorf("unknown template %s, can be one of ping, excuses", i.template)
	}
	log.Info("Creating repo from template...")
	var githubClient *git.GithubClient
	var err error
	if !i.noRepo {
		githubClient, err = git.NewGithubClient(i.githubOrg)
		if err != nil {
			return fmt.Errorf("could not initialize github client - %v", err)
		}
	}
	lc := mantil.LocalConfig(i.name, i.githubOrg)
	repoURL, err := git.CreateRepoFromTemplate(templateRepo, i.name, i.name, githubClient, lc)
	if err != nil {
		return fmt.Errorf("could not create repo %s from template - %v", i.name, err)
	}
	token, err := i.initRequest(i.name)
	if err != nil || token == "" {
		return fmt.Errorf("could not initialize project - %v", err)
	}
	if repoURL != "" {
		backendURL, err := commands.BackendURL()
		if err != nil {
			return fmt.Errorf("could not get backedn URL - %v", err)
		}
		if err := githubClient.AddSecrets(i.name, token, backendURL); err != nil {
			return fmt.Errorf("could not add mantil token to repo - %v", err)
		}
	}
	if err := mantil.SaveToken(i.name, token); err != nil {
		return fmt.Errorf("could not save token to ~/.mantil directory - %v", err)
	}
	projectPath, _ := filepath.Abs(i.name)
	log.Notice("Done!\nProject initialized at %s", projectPath)
	if repoURL != "" {
		log.Notice("Github repo URL: %s", repoURL)
	}
	return nil
}

func (i *InitCmd) templateRepo() string {
	ping := "https://github.com/mantil-io/go-mantil-template"
	switch i.template {
	case "excuses":
		return "https://github.com/mantil-io/template-excuses"
	case "ping":
		return ping
	case "":
		return ping
	}
	return ""
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
