package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/git"
	"github.com/mantil-io/mantil/workspace"
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
	ui.Info("Cloning into %s and replacing import paths with %s", projectPath, c.moduleName)
	if err := git.CreateRepo(repo, c.name, c.moduleName); err != nil {
		return fmt.Errorf("could not clone %s - %v", repo, err)
	}
	project := &workspace.Project{
		Name: c.name,
	}
	if err := workspace.SaveProject(project, projectPath); err != nil {
		return err
	}
	ui.Notice("Done!")
	ui.Notice("Project initialized at %s", projectPath)
	return nil
}

func (c *newCmd) repoURL() (string, error) {
	repo := c.repo
	if c.isExternalRepo() {
		ui.Info("Creating project %s from external repository %s...", c.name, repo)
	} else {
		template := c.template()
		if template == "" {
			return "", fmt.Errorf("project source recognised as template but it's not one of valid values, can be one of: ping, excuses")
		}
		repo = c.templateRepo(template)
		ui.Info("Creating project %s from template %s...", c.name, template)
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
