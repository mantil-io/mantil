package initialize

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/mantil/commands"
	"github.com/mantil-io/mantil/cli/mantil/log"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/git"
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
	repo, err := i.repoURL()
	if err != nil {
		return err
	}
	log.Info("Cloning into %s and replacing import paths with %s", projectPath, i.moduleName)
	if err := git.CreateRepo(repo, i.name, i.moduleName); err != nil {
		return fmt.Errorf("could not clone %s - %v", repo, err)
	}
	project, err := i.initRequest(i.name)
	if err != nil {
		return err
	}
	if err := config.SaveProject(project, projectPath); err != nil {
		return err
	}
	log.Notice("Done!")
	log.Notice("Project initialized at %s", projectPath)
	return nil
}

func (i *InitCmd) repoURL() (string, error) {
	repo := i.repo
	if i.isExternalRepo() {
		log.Info("Creating project %s from external repository %s...", i.name, repo)
	} else {
		template := i.template()
		if template == "" {
			return "", fmt.Errorf("project source recognised as template but it's not one of valid values, can be one of: ping, excuses")
		}
		repo = i.templateRepo(template)
		log.Info("Creating project %s from template %s...", i.name, template)
	}
	return repo, nil
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

func (i *InitCmd) initRequest(projectName string) (*config.Project, error) {
	type initReq struct {
		ProjectName string
	}
	ireq := &initReq{
		ProjectName: projectName,
	}
	type initResp struct {
		Project *config.Project
	}
	iresp := &initResp{}
	if err := commands.BackendRequest("init", ireq, iresp, true); err != nil {
		return nil, err
	}
	return iresp.Project, nil
}