package term

import (
	"os"

	"golang.org/x/term"
)

func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
