package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/commands"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/git"
)

type newCmd struct {
	name       string
	repo       string
	moduleName string
}

func (c *newCmd) run() error {
	projectPath, _ := filepath.Abs(c.name)
	repo, err := c.repoURL()
	if err != nil {
		return err
	}
	log.Info("Cloning into %s and replacing import paths with %s", projectPath, c.moduleName)
	if err := git.CreateRepo(repo, c.name, c.moduleName); err != nil {
		return fmt.Errorf("could not clone %s - %v", repo, err)
	}
	project, err := c.newRequest(c.name)
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

func (c *newCmd) repoURL() (string, error) {
	repo := c.repo
	if c.isExternalRepo() {
		log.Info("Creating project %s from external repository %s...", c.name, repo)
	} else {
		template := c.template()
		if template == "" {
			return "", fmt.Errorf("project source recognised as template but it's not one of valid values, can be one of: ping, excuses")
		}
		repo = c.templateRepo(template)
		log.Info("Creating project %s from template %s...", c.name, template)
	}
	return repo, nil
}

func (c *newCmd) isExternalRepo() bool {
	return strings.HasPrefix(c.repo, "http") || strings.HasPrefix(c.repo, "git")
}

func (c *newCmd) template() string {
	switch c.repo {
	case "excuses", "ping":
		return c.repo
	case "":
		return "ping"
	}
	return ""
}

func (c *newCmd) templateRepo(template string) string {
	switch template {
	case "excuses":
		return "https://github.com/mantil-io/template-excuses"
	case "ping":
		return "https://github.com/mantil-io/go-mantil-template"
	}
	return ""
}

func (c *newCmd) newRequest(projectName string) (*config.Project, error) {
	type newReq struct {
		ProjectName string
	}
	ireq := &newReq{
		ProjectName: projectName,
	}
	type newResp struct {
		Project *config.Project
	}
	iresp := &newResp{}
	if err := commands.BackendRequest("init", ireq, iresp, true); err != nil {
		return nil, err
	}
	return iresp.Project, nil
}
