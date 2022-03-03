package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mantil-io/mantil/cli/controller"
	"github.com/mantil-io/mantil/cli/controller/examples"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/spf13/cobra"
)

func Execute() error {
	ec, err := root().ExecuteC()
	defer controller.Defer()
	if err == nil {
		return nil
	}

	// show usage in case of ArgumentError
	var ae *controller.ArgumentError
	if errors.As(err, &ae) {
		ui.Error(ae)
		ec.Usage()
		return err
	}

	// if the error is not wrapped then it is probably generated by cobra
	// show usage because that is breaking some arguments constraints
	if errors.Unwrap(err) == nil {
		ui.Error(err)
		ec.Usage()
		return err
	}

	showError(ec, err)
	// in other cases show error without usage
	//ui.Error(err)
	return err
}

func root() *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "mantil",
		Short:         "Makes serverless development with Go and AWS Lambda joyful.",
		Version:       domain.Version(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// no-color is handled in cli/log pacakge
	cmd.PersistentFlags().Bool("no-color", false, "Don't use colors in output")
	cmd.PersistentFlags().Bool("help", false, "Show command help") // move help to global commands
	cmd.Flags().Bool("version", false, "Show mantil version")      // remove -v shortcut for version
	cmd.SetUsageTemplate(usageTemplate(""))

	add := func(factory func() *cobra.Command) {
		sub := factory()
		addCommand(cmd, sub)
	}
	subCommands := []func() *cobra.Command{
		newEnvCommand,
		newInvokeCommand,
		newLogsCommand,
		newNewCommand,
		newTestCommand,
		newWatchCommand,
		newDeployCommand,
		newGenerateCommand,
		newAwsCommand,
		newStageCommand,
		newReportCommand,

		// for testing:
		//examples.NewErrorsCommand,
		//examples.NewArgsCommand,
		examples.NewGithubAuthCommand,
		examples.NewNodeCommand,
	}
	for _, sub := range subCommands {
		add(sub)
	}
	return cmd
}

func GenDoc(dir string) error {
	if err := (&mdGenerator{dir: dir}).gen(root()); err != nil {
		ui.Error(err)
	}
	return nil
}

func usageTemplate(argumentsUsage string) string {
	if argumentsUsage != "" {
		argumentsUsage = fmt.Sprintf(`

\bARGUMENTS\c%s`, argumentsUsage)
	}

	str := fmt.Sprintf(`\bUSAGE\c{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

\bALIASES\c
  {{.NameAndAliases}}{{end}}{{if .HasAvailableSubCommands}}

\bCOMMANDS\c{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

  Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}%s{{if .HasAvailableLocalFlags}}

\bOPTIONS\c
{{.LocalFlags.FlagUsagesWrapped 120 | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

\bEXAMPLES\c{{.Example}}{{end}}{{if .HasAvailableInheritedFlags}}

\bGLOBAL OPTIONS\c
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

\bADDITIONAL HELP TOPICS\C{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}

\bLEARN MORE\c
  Visit https://github.com/mantil-io/docs to learn more.
  For further support contact us at support@mantil.com.

`, argumentsUsage)

	return boldize(str)
}

func setUsageTemplate(cmd *cobra.Command, argumentsUsage string) {
	cmd.SetUsageTemplate(usageTemplate(argumentsUsage))
}

func boldize(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(str,
		`\b`, bold),
		`\c`, clear)
}

const (
	bold  = "\033[1m"
	clear = "\033[0m"
)

func showError(cmd *cobra.Command, err error) {
	if err == nil {
		return
	}

	var nee *domain.NodeExistsError
	if errors.As(err, &nee) {
		ui.Errorf("node '%s' already exists", nee.Name)
		if nee.Name == domain.DefaultNodeName {
			fmt.Printf(`
'%s' is default node name and it is already used.
Please specify another name in mantil command.
`, nee.Name)
		}
		return
	}

	var verr *domain.ValidationError
	if errors.As(err, &verr) {
		ui.Errorf("%s is not a valid name", verr.Name)
		fmt.Print(`
Please check the following rules when naming projects, stages and functions:
- Names cannot be longer than 16 characters
- Names can only contain letters, numbers and the special characters - and _
`)
		return
	}

	var gbe *log.GoBuildError
	if errors.As(err, &gbe) {
		for _, line := range gbe.Lines {
			ui.ErrorLine(line)
		}
		return
	}

	var pnse *domain.ProjectNoStagesError
	if errors.As(err, &pnse) {
		ui.Info("This project doesn't have active stages.")
		ui.Info("You can create new stage with 'mantil stage new'.")
		return
	}

	var wnne *domain.WorkspaceNoNodesError
	if errors.As(err, &wnne) {
		ui.Info("There are no nodes installed in your workspace.")
		ui.Info("To install new node please run 'mantil aws install'.")
		return
	}

	var nnfe *domain.NodeNotFoundError
	if errors.As(err, &nnfe) {
		ui.Info("Node %s doesn't exist. For a complete list of available nodes run 'mantil aws ls'.", nnfe.Name)
		return
	}

	var perr *domain.ProjectNotFoundError
	if errors.As(err, &perr) {
		ui.Info("Mantil project was not found in path. This command needs to be run inside project structure.")
		store, err := domain.NewSingleDeveloperWorkspaceStore()
		if err != nil {
			return
		}
		if len(store.Workspace().Projects) == 0 {
			ui.Info("You can create new project with 'mantil new'.")
			return
		}
		ui.Info("\nCurrent projects with active stages:")
		var data [][]string
		for _, p := range store.Workspace().Projects {
			data = append(data, []string{p.Name, p.Path})
		}
		controller.ShowTable([]string{"name", "path"}, data)
		return
	}

	var naud *domain.NodeAlreadyUpToDateError
	if errors.As(err, &naud) {
		ui.Info("Node %s is already up to date with the current version %s of CLI.", naud.Name, naud.Version)
		return
	}

	var ane *controller.ApiNewError
	if errors.As(err, &ane) {
		ui.Errorf("function New for api %s does not have proper type", ane.Api)
		ui.Info("Function should have no parameters and only one return value of type struct or pointer to struct.")
		return
	}

	ui.Error(err)
}

func addCommand(target, cmd *cobra.Command) {
	cmd.DisableFlagsInUseLine = true
	if cmd.HasAvailableFlags() && !strings.Contains(cmd.Use, "[options]") {
		cmd.Use += " [options]"
	}
	target.AddCommand(cmd)
}
