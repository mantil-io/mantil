package examples

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func NewArgsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "args",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	cmd.AddCommand(NewArgsSimpleCommand())
	cmd.AddCommand(NewArgsComplexCommand())
	return cmd
}

func NewArgsSimpleCommand() *cobra.Command {
	// flags definition
	var force bool

	// command description
	cmd := &cobra.Command{
		Use:   "simple [attr]",
		Short: "some short description",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("force: %v\n", force)
			fmt.Printf("attr: %s", args[0])
			return nil
		},
	}

	// flags binding
	cmd.Flags().BoolVar(&force, "force", false, "example of bool option")
	return cmd
}

func NewArgsComplexCommand() *cobra.Command {
	// args definition
	a := struct {
		str   string
		force bool
		since time.Duration
		str2  *string
		arg0  string
	}{}

	// command description
	cmd := &cobra.Command{
		Use:     "complex [name]",
		Short:   "some short description",
		Args:    cobra.MaximumNArgs(1),
		Example: "mantil flags complex --str jozo --force --since 12h4m3s --str2 pero 'arg0 value'",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				a.arg0 = args[0]
			}
			// call something which uses just args
			fmt.Printf("args: %#v\n", a)
			fmt.Printf("str2: %v %s\n", a.str2, *a.str2)
			return nil
		},
	}

	// flags binding
	cmd.Flags().BoolVar(&a.force, "force", false, "example of bool option")
	cmd.Flags().StringVar(&a.str, "str", "", "some string option")
	cmd.Flags().DurationVar(&a.since, "since", 3*time.Hour, "duration... why use this")
	// ne vidim prednosti ovog nacina (samo nedostatke) pa ga obeshrabrujem
	// nedostatak je sto imas pointer, koji je uvijek non nil, pa sto ce mi onda pointer kada ne nosi nikakvu novu informaciju
	// a onda ga dalje uvijek moram dereferencirati
	a.str2 = cmd.Flags().String("str2", "", "some other string option")
	return cmd
}
