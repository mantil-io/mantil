package controller

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/git"
)

const (
	PingTemplate    = "ping"
	ExcusesTemplate = "excuses"
	ChatTemplate    = "chat"
	TodoTemplate    = "todo"
	G2STemplate     = "github-to-slack"
)

var TemplateRepos = map[string]string{
	PingTemplate:    "https://github.com/mantil-io/template-ping",
	ExcusesTemplate: "https://github.com/mantil-io/template-excuses",
	ChatTemplate:    "https://github.com/mantil-io/template-chat",
	TodoTemplate:    "https://github.com/mantil-io/template-todo",
	G2STemplate:     "https://github.com/mantil-io/template-github-to-slack",
}

const (
	DefaultTemplate = PingTemplate
	LicenseFile     = "LICENSE"
	ReadmeFile      = "README.md"
	ImagesDir       = "images"
)

type NewArgs struct {
	Name       string
	From       string
	ModuleName string
}

func New(a NewArgs) error {
	if a.ModuleName == "" {
		a.ModuleName = a.Name
	}
	if a.From == "" {
		a.From = DefaultTemplate
	}
	return createProject(a.Name, a.From, a.ModuleName)
}

func createProject(name, from, moduleName string) error {
	if err := domain.ValidateName(name); err != nil {
		return log.Wrap(err)
	}
	projectPath, _ := filepath.Abs(name)
	repo, err := repoURL(name, from)
	if err != nil {
		return log.Wrap(err)
	}
	ui.Info("")
	ui.Info("Creating %s in %s...", name, projectPath)
	ui.Info("Replacing import paths with %s...", moduleName)
	if err := git.CreateRepo(repo, name, moduleName); err != nil {
		if errors.Is(err, git.ErrRepositoryNotFound) {
			return log.Wrap(err, sourceNewUserError(repo))
		}
		return log.Wrap(err, "could not initialize repository from source %s: %v", repo, err)

	}

	// delete unnecessary files from template repositories
	if !isExternalRepo(from) {
		os.Remove(filepath.Join(projectPath, LicenseFile))
		os.Remove(filepath.Join(projectPath, ReadmeFile))
		os.RemoveAll(filepath.Join(projectPath, ImagesDir))
	}

	fs, err := newStore()
	if err != nil {
		return log.Wrap(err)
	}
	if err := fs.NewProject(name, projectPath); err != nil {
		return log.Wrap(err)
	}
	log.Event(domain.Event{ProjectNew: &domain.ProjectNew{
		Name: name,
		From: from,
		Repo: repo,
	}})
	ui.Info("")
	ui.Title("Your project is ready in %s\n", projectPath)
	return nil
}

func repoURL(name, repo string) (string, error) {
	if !isExternalRepo(repo) {
		template := projectTemplate(repo)
		if template == "" {
			return "", log.Wrapf("invalid template %s", repo)

		}
		repo = TemplateRepos[template]
	}
	return repo, nil
}

func isExternalRepo(repo string) bool {
	return strings.HasPrefix(repo, "http") || strings.HasPrefix(repo, "git@")
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
