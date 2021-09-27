package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil/cli/commands"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/git"
	"github.com/spf13/cobra"
)

func init() {
	addCommandNew(rootCmd)
}

func addCommandNew(rootCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "new <project>",
		Short: "Initializes a new Mantil project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			RunNew(cmd, args)
		},
	}
	cmd.Flags().String("from", "", "name of the template or URL of the repository that will be used as one")
	cmd.Flags().String("module-name", "", "replace module name and import paths")
	rootCmd.AddCommand(cmd)
}

func RunNew(cmd *cobra.Command, args []string) {
	nc := initNew(cmd, args)
	if err := nc.newProject(); err != nil {
		log.Fatal(err)
	}
}

type newCmd struct {
	name       string
	repo       string
	moduleName string
}

func initNew(cmd *cobra.Command, args []string) *newCmd {
	projectName := args[0]
	repo := cmd.Flag("from").Value.String()
	moduleName := cmd.Flag("module-name").Value.String()
	if moduleName == "" {
		moduleName = projectName
	}
	return &newCmd{
		name:       projectName,
		repo:       repo,
		moduleName: moduleName,
	}
}

func (n *newCmd) newProject() error {
	projectPath, _ := filepath.Abs(n.name)
	repo, err := n.repoURL()
	if err != nil {
		return err
	}
	log.Info("Cloning into %s and replacing import paths with %s", projectPath, n.moduleName)
	if err := git.CreateRepo(repo, n.name, n.moduleName); err != nil {
		return fmt.Errorf("could not clone %s - %v", repo, err)
	}
	project, err := n.newRequest(n.name)
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

func (n *newCmd) repoURL() (string, error) {
	repo := n.repo
	if n.isExternalRepo() {
		log.Info("Creating project %s from external repository %s...", n.name, repo)
	} else {
		template := n.template()
		if template == "" {
			return "", fmt.Errorf("project source recognised as template but it's not one of valid values, can be one of: ping, excuses")
		}
		repo = n.templateRepo(template)
		log.Info("Creating project %s from template %s...", n.name, template)
	}
	return repo, nil
}

func (n *newCmd) isExternalRepo() bool {
	return strings.HasPrefix(n.repo, "http") || strings.HasPrefix(n.repo, "git")
}

func (n *newCmd) template() string {
	switch n.repo {
	case "excuses", "ping":
		return n.repo
	case "":
		return "ping"
	}
	return ""
}

func (n *newCmd) templateRepo(template string) string {
	switch template {
	case "excuses":
		return "https://github.com/mantil-io/template-excuses"
	case "ping":
		return "https://github.com/mantil-io/go-mantil-template"
	}
	return ""
}

func (n *newCmd) newRequest(projectName string) (*config.Project, error) {
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
