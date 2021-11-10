package progress

import (
	"bytes"
	"io"
	"os"

	"golang.org/x/term"
)

type writer struct {
	out            io.Writer
	buffer         *bytes.Buffer
	prevLineCount  int
	isTerminal     bool
	terminalWidth  int
	terminalHeight int
}

func newWriter(out io.Writer) *writer {
	if out == nil {
		out = os.Stdout
	}
	w := &writer{
		out:    out,
		buffer: bytes.NewBuffer(nil),
	}
	w.initTerminal()
	return w
}

func (w *writer) initTerminal() {
	fd := int(os.Stdout.Fd())
	w.isTerminal = term.IsTerminal(fd)
	width, height, _ := term.GetSize(fd)
	w.terminalWidth = width
	w.terminalHeight = height
}

// Write to internal buffer, won't be visible until flush() is called
func (w *writer) Write(p []byte) (int, error) {
	n, err := w.buffer.Write(p)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// flush all buffered changes and reset internal buffer
func (w *writer) flush() error {
	// add an additional newline so that we can move the cursor up
	// without shifting the whole output on every flush
	w.buffer.WriteByte('\n')
	lc := w.lineCount()
	w.clearLines(w.prevLineCount)
	w.prevLineCount = lc
	if _, err := w.out.Write(w.buffer.Bytes()); err != nil {
		return err
	}
	w.buffer.Reset()
	return nil
}

func (w *writer) clearLines(n int) error {
	for i := 0; i < n; i++ {
		if err := w.clearLine(); err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) lineCount() int {
	w.initTerminal()
	if !w.isTerminal {
		return 1
	}
	lines := 0
	currentCnt := 0
	for _, b := range w.buffer.Bytes() {
		currentCnt++
		if b == '\n' || currentCnt > w.terminalWidth {
			lines++
			currentCnt = 0
		}
	}
	return lines
}
