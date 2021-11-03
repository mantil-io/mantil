package controller

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/git"
	"github.com/pkg/errors"
)

var TemplateRepos = map[string]string{
	"ping":    "https://github.com/mantil-io/go-mantil-template",
	"excuses": "https://github.com/mantil-io/template-excuses",
}

const DefaultTemplate = "ping"

type NewArgs struct {
	Name       string
	Repo       string
	ModuleName string
}

func New(a NewArgs) error {
	if a.ModuleName == "" {
		a.ModuleName = a.Name
	}
	if a.Repo == "" {
		a.Repo = DefaultTemplate
	}
	return createProject(a.Name, a.Repo, a.ModuleName)
}

func createProject(name, repo, moduleName string) error {
	if err := domain.ValidateName(name); err != nil {
		return log.Wrap(err)
	}
	projectPath, _ := filepath.Abs(name)
	repo, err := repoURL(name, repo)
	if err != nil {
		return log.Wrap(err)
	}
	ui.Info("Cloning into %s and replacing import paths with %s...", projectPath, moduleName)
	if err := git.CreateRepo(repo, name, moduleName); err != nil {
		if errors.Is(err, git.ErrRepositoryNotFound) {
			return log.Wrap(err, sourceNewUserError(repo))
		}
		return log.Wrap(err, "Could not initialize repository from source %s: %v", repo, err)

	}
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	if err != nil {
		return log.Wrap(err)
	}
	if err := fs.NewProject(name, projectPath); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Project initialized in %s", projectPath)
	return nil
}

func repoURL(name, repo string) (string, error) {
	if isExternalRepo(repo) {
		ui.Info("Creating project %s from external repository %s...", name, repo)
	} else {
		template := projectTemplate(repo)
		if template == "" {
			return "", log.Wrap(fmt.Errorf("invalid template %s", repo))

		}
		repo = TemplateRepos[template]
		ui.Info("Creating project %s from template %s...", name, template)
	}
	return repo, nil
}

func isExternalRepo(repo string) bool {
	return strings.HasPrefix(repo, "http") || strings.HasPrefix(repo, "git")
}

func projectTemplate(repo string) string {
	if repo == "" {
		return DefaultTemplate
	}
	if _, exists := TemplateRepos[repo]; exists {
		return repo
	}
	return ""
}

func sourceNewUserError(repo string) string {
	var validTemplates []string
	for t := range TemplateRepos {
		validTemplates = append(validTemplates, t)
	}
	return fmt.Sprintf(`%s is not a valid project source, please provide one of the following:
- a link to an existing git repository, starting with 'http(s):' or 'git:'
- the name of a predefined template which can be one of: %s`, repo, strings.Join(validTemplates, ", "))
}
