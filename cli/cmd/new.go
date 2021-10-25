package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/git"
	"github.com/mantil-io/mantil/workspace"
	"github.com/pkg/errors"
)

var templateRepos = map[string]string{
	"ping":    "https://github.com/mantil-io/go-mantil-template",
	"excuses": "https://github.com/mantil-io/template-excuses",
}

const defaultTemplate = "ping"

type newArgs struct {
	name       string
	repo       string
	moduleName string
}

type newCmd struct {
	name       string
	repo       string
	moduleName string
}

func newNew(a newArgs) (*newCmd, error) {
	if a.moduleName == "" {
		a.moduleName = a.name
	}
	if a.repo == "" {
		a.repo = defaultTemplate
	}
	return &newCmd{
		name:       a.name,
		repo:       a.repo,
		moduleName: a.moduleName,
	}, nil
}

func (c *newCmd) run() error {
	if err := workspace.ValidateName(c.name); err != nil {
		return log.Wrap(err)
	}
	projectPath, _ := filepath.Abs(c.name)
	repo, err := c.repoURL()
	if err != nil {
		return log.Wrap(err)
	}
	ui.Info("Cloning into %s and replacing import paths with %s...", projectPath, c.moduleName)
	if err := git.CreateRepo(repo, c.name, c.moduleName); err != nil {
		if errors.Is(err, git.ErrRepositoryNotFound) {
			return log.Wrap(err, c.sourceUserError())
		}
		return log.Wrap(err, "Could not initialize repository from source %s: %v", repo, err)

	}
	fs, err := workspace.NewSingleDeveloperWorkspaceStore()
	if err != nil {
		return log.Wrap(err)
	}
	if err := fs.NewProject(c.name, projectPath); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Project initialized in %s", projectPath)
	return nil
}

func (c *newCmd) repoURL() (string, error) {
	repo := c.repo
	if c.isExternalRepo() {
		ui.Info("Creating project %s from external repository %s...", c.name, repo)
	} else {
		template := c.template()
		if template == "" {
			return "", log.Wrap(fmt.Errorf("invalid template %s", c.repo))

		}
		repo = templateRepos[template]
		ui.Info("Creating project %s from template %s...", c.name, template)
	}
	return repo, nil
}

func (c *newCmd) sourceUserError() string {
	var validTemplates []string
	for t := range templateRepos {
		validTemplates = append(validTemplates, t)
	}
	return fmt.Sprintf(`%s is not a valid project source, please provide one of the following:
- a link to an existing git repository, starting with 'http(s):' or 'git:'
- the name of a predefined template which can be one of: %s`, c.repo, strings.Join(validTemplates, ", "))
}

func (c *newCmd) isExternalRepo() bool {
	return strings.HasPrefix(c.repo, "http") || strings.HasPrefix(c.repo, "git")
}

func (c *newCmd) template() string {
	if c.repo == "" {
		return defaultTemplate
	}
	if _, exists := templateRepos[c.repo]; exists {
		return c.repo
	}
	return ""
}
