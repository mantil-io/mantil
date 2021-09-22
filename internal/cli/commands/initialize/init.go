package initialize

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/internal/cli/commands"
	"github.com/mantil-io/mantil/internal/cli/git"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/mantil"
)

type InitCmd struct {
	name       string
	repo       string
	moduleName string
}

func New(name, repo, moduleName string) (*InitCmd, error) {
	return &InitCmd{
		name:       name,
		repo:       repo,
		moduleName: moduleName,
	}, nil
}

func (i *InitCmd) InitProject() error {
	projectPath, _ := filepath.Abs(i.name)
	repo, isTemplate, err := i.repoURL()
	if err != nil {
		return err
	}
	log.Info("Cloning into %s...", projectPath)
	if err := git.CreateRepo(repo, i.name, i.moduleName, isTemplate); err != nil {
		return fmt.Errorf("could not clone %s - %v", repo, err)
	}
	token, err := i.initRequest(i.name)
	if err != nil || token == "" {
		return fmt.Errorf("could not initialize project - %v", err)
	}
	if _, err := mantil.CreateLocalConfig(i.name); err != nil {
		return fmt.Errorf("could not create local project config - %v", err)
	}
	if err := mantil.SaveToken(i.name, token); err != nil {
		return fmt.Errorf("could not save token to ~/.mantil directory - %v", err)
	}
	log.Notice("Done!")
	log.Notice("Project initialized at %s", projectPath)
	return nil
}

func (i *InitCmd) repoURL() (string, bool, error) {
	repo := i.repo
	isTemplate := false
	if i.isExternalRepo() {
		log.Info("Creating project %s from external repository %s...", i.name, repo)
	} else {
		template := i.template()
		if template == "" {
			return "", false, fmt.Errorf("project source recognised as template but it's not one of valid values, can be one of: ping, excuses")
		}
		repo = i.templateRepo(template)
		isTemplate = true
		log.Info("Creating project %s from template %s...", i.name, template)
	}
	return repo, isTemplate, nil
}

func (i *InitCmd) isExternalRepo() bool {
	return strings.HasPrefix(i.repo, "http") || strings.HasPrefix(i.repo, "git")
}

func (i *InitCmd) template() string {
	switch i.repo {
	case "excuses", "ping":
		return i.repo
	case "":
		return "ping"
	}
	return ""
}

func (i *InitCmd) templateRepo(template string) string {
	switch template {
	case "excuses":
		return "https://github.com/mantil-io/template-excuses"
	case "ping":
		return "https://github.com/mantil-io/go-mantil-template"
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
