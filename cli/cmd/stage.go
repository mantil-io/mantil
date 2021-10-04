package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/cmd/deploy"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/git"
	"github.com/mantil-io/mantil/workspace"
	"github.com/spf13/cobra"
)

func newStageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stage",
		Short: "Manage project stages",
	}
	cmd.AddCommand(newStageNewCommand())
	cmd.AddCommand(newStageDestroyCommand())
	return cmd
}

func newStageNewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new <name>",
		Short: "Create a new stage",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			accountName, _ := cmd.Flags().GetString("account")
			return initStageCommand(args).new(accountName)
		},
	}
	cmd.Flags().StringP("account", "a", "", "account in which the stage will be created")
	return cmd
}

func newStageDestroyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destroy <name>",
		Short: "Destroy a stage",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, _ := cmd.Flags().GetBool("delete-repo")
			force, _ := cmd.Flags().GetBool("force")
			all, _ := cmd.Flags().GetBool("all")
			return initStageCommand(args).destroy(repo, force, all)
		},
	}
	cmd.Flags().Bool("delete-repo", false, "delete local repository")
	cmd.Flags().Bool("force", false, "don't ask for confirmation")
	cmd.Flags().Bool("all", false, "destroy all stages")
	return cmd
}

type stageCmd struct {
	stageName string
	ctx       *project.Context
}

func initStageCommand(args []string) *stageCmd {
	var stageName string
	if len(args) == 1 {
		stageName = args[0]
	}
	ctx := project.MustContext()
	return &stageCmd{
		stageName: stageName,
		ctx:       ctx,
	}
}

func (c *stageCmd) new(accountName string) error {
	if c.stageName == "" {
		c.stageName = workspace.DefaultStageName
	}
	if s := c.ctx.Project.Stage(c.stageName); s != nil {
		return log.WithUserMessage(nil, fmt.Sprintf("Stage %s already exists.", c.stageName))
	}
	ui.Info("Creating stage %s...", c.stageName)
	stage, err := c.createStage(accountName)
	if err != nil {
		return err
	}
	if err = c.ctx.SetStage(stage); err != nil {
		return err
	}
	aws := c.ctx.MustInitialiseAWSSDK()
	d, err := deploy.New(c.ctx, aws)
	if err != nil {
		return err
	}
	_, err = d.Deploy()
	if err != nil {
		return err
	}
	return nil
}

func (c *stageCmd) createStage(accountName string) (*workspace.Stage, error) {
	if len(c.ctx.Workspace.Accounts) == 0 {
		return nil, log.WithUserMessage(nil, "No accounts found in workspace. Please add an account with `mantil install`.")
	}
	if accountName == "" {
		if len(c.ctx.Workspace.Accounts) > 1 {
			accountName = c.selectAccount()
		} else {
			accountName = c.ctx.Workspace.Accounts[0].Name
		}
	}
	stage := &workspace.Stage{
		Name:    c.stageName,
		Account: accountName,
	}
	if len(c.ctx.Project.Stages) == 0 {
		stage.Default = true
	}
	return stage, nil
}

func (s *stageCmd) selectAccount() string {
	var accounts []string
	for _, a := range s.ctx.Workspace.Accounts {
		accounts = append(accounts, a.Name)
	}
	prompt := promptui.Select{
		Label: "Select an account",
		Items: accounts,
	}
	_, account, err := prompt.Run()
	if err != nil {
		ui.Fatal(err)
	}
	return account
}

func (c *stageCmd) destroy(deleteRepo, force, destroyAll bool) error {
	if !destroyAll && c.stageName == "" {
		return log.WithUserMessage(nil, "No stage specified")
	}
	if !force {
		c.confirmDestroy(destroyAll)
	}
	if destroyAll {
		for _, s := range c.ctx.Project.Stages {
			if err := c.destroyStage(s); err != nil {
				return err
			}
		}
	} else {
		s := c.ctx.Project.Stage(c.stageName)
		if s == nil {
			return fmt.Errorf("stage %s not found", c.stageName)
		}
		if err := c.destroyStage(s); err != nil {
			return err
		}
	}
	if deleteRepo {
		ui.Info("Deleting local repository...")
		if err := git.DeleteRepo(c.ctx.Path); err != nil {
			return err
		}
	}
	c.ctx.Project.SetDefaultStage()
	workspace.SaveProject(c.ctx.Project, c.ctx.Path)
	ui.Notice("Destroy successfully finished")
	return nil
}

func (c *stageCmd) confirmDestroy(destroyAll bool) {
	var label string
	if destroyAll {
		label = "To confirm deletion of all stages, please enter the project name"
	} else {
		label = fmt.Sprintf("To confirm deletion of stage %s, please enter the project name", c.stageName)
	}
	confirmationPrompt := promptui.Prompt{
		Label: label,
	}
	projectName, err := confirmationPrompt.Run()
	if err != nil {
		ui.Fatal(err)
	}
	if c.ctx.Project.Name != projectName {
		ui.Fatalf("Project name doesn't match")
	}
}

func (c *stageCmd) destroyStage(stage *workspace.Stage) error {
	if err := c.ctx.SetStage(stage); err != nil {
		return err
	}
	ui.Info("Destroying stage %s in account %s", c.ctx.Stage.Name, c.ctx.Account.Name)
	if err := c.destroyRequest(); err != nil {
		return fmt.Errorf("could not destroy stage %s - %v", c.ctx.Stage.Name, err)
	}
	c.ctx.Project.RemoveStage(c.ctx.Stage.Name)
	return nil
}

func (c *stageCmd) destroyRequest() error {
	type req struct {
		ProjectName string
		StageName   string
	}
	r := &req{
		ProjectName: c.ctx.Project.Name,
		StageName:   c.ctx.Stage.Name,
	}
	if err := c.ctx.RuntimeRequest("destroy", r, nil, true); err != nil {
		return err
	}
	return nil
}
