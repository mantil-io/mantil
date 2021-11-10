package progress

import (
	"bytes"
	"io"
	"os"
)

type writer struct {
	out       io.Writer
	buffer    *bytes.Buffer
	bufferLen int // keep track of the buffer length, this is useful for the windows implementation
}

func newWriter(out io.Writer) *writer {
	if out == nil {
		out = os.Stdout
	}
	w := &writer{
		out:    out,
		buffer: bytes.NewBuffer(nil),
	}
	w.reset()
	return w
}

// Write to internal buffer, won't be visible until flush() is called
func (w *writer) Write(p []byte) (int, error) {
	n, err := w.buffer.Write(p)
	if err != nil {
		return 0, err
	}
	w.bufferLen += n
	return n, nil
}

func (w *writer) reset() error {
	w.buffer.Reset()
	// add the clearLine string to the empty buffer, this will clear the previous
	// line when flushed out so we don't have to do two writes later
	return w.clearLine()
}

// flush all buffered changes and reset internal buffer
func (w *writer) flush() error {
	if _, err := w.out.Write(w.buffer.Bytes()); err != nil {
		return err
	}
	if err := w.reset(); err != nil {
		return err
	}
	w.bufferLen = 0
	return nil
}
