//go:build windows

package progress

import (
	"strings"
)

func (w *writer) clearLine() error {
	clr := "\r"
	// overwrite the previous line with spaces and return to the start
	clr = strings.Repeat(" ", w.bufferLen)
	clr += "\r"
	_, err := w.buffer.Write([]byte(clr))
	return err
}
