package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mantil-io/mantil/cli/build"
	"github.com/mantil-io/mantil/cli/controller"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func Execute() error {
	ec, err := root().ExecuteC()
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
		Short:         "Makes serverless development with Go and AWS Lambda joyful",
		Version:       build.Version(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	// no-color is handled in cli/log pacakge
	cmd.PersistentFlags().Bool("no-color", false, "don't use colors in output")
	cmd.PersistentFlags().Bool("help", false, "show command help") // move help to global commands
	cmd.Flags().Bool("version", false, "show mantil version")      // remove -v shortcut for version
	cmd.SetUsageTemplate(usageTemplate(""))

	add := func(factory func() *cobra.Command) {
		sub := factory()
		cmd.AddCommand(sub)
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

		// for testing:
		//examples.NewErrorsCommand,
		//examples.NewArgsCommand,
	}
	for _, sub := range subCommands {
		add(sub)
	}
	return cmd
}

func GenDoc(dir string) error {
	cmd := root()
	cmd.DisableAutoGenTag = true
	return doc.GenMarkdownTree(cmd, dir)
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

\bAVAILABLE COMMANDS\c{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}{{if .HasAvailableLocalFlags}}
%s
\bFLAGS\c
{{.LocalFlags.FlagUsagesWrapped 120 | trimTrailingWhitespaces}}{{end}}{{if .HasExample}}

\bEXAMPLES\c
{{.Example}}{{end}}{{if .HasAvailableInheritedFlags}}

\bGLOBAL FLAGS\c
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

\bADDITIONAL HELP TOPICS\C{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}

\bLEARN MORE\c
  Visit https://docs.mantil.io to learn more.
  For further support contact us at hello@mantil.com.
`, argumentsUsage)

	return boldize(str)
}

func setUsageTemplate(cmd *cobra.Command, argumentsUsage string) {
	cmd.SetUsageTemplate(usageTemplate(argumentsUsage))
}

func showNextSteps(str string) {
	fmt.Printf("==> Next steps:%s\n", str)
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

	var aee *domain.AccountExistsError
	if errors.As(err, &aee) {
		ui.Errorf("account '%s' already exists", aee.Name)
		if aee.Name == domain.DefaultAccountName {
			fmt.Printf(`
'%s' is default account name and it is already used.
Please specify another name in mantil command.
`, aee.Name)
		}
		return
	}

	var rerr *domain.ErrReservedName
	if errors.As(err, &rerr) {
		ui.Errorf("'%s' is reserved name", rerr.Name)
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
	//ok, _ := cmd.InheritedFlags().GetBool("no-color")

	ui.Error(err)
}
