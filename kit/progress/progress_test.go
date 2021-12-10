package progress

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

type testWriter struct {
	buf   *bytes.Buffer
	outCh chan string
}

func newTestWriter() *testWriter {
	return &testWriter{
		buf:   bytes.NewBuffer(nil),
		outCh: make(chan string),
	}
}

func (tw *testWriter) Write(p []byte) (int, error) {
	return tw.buf.Write(p)
}

func (tw *testWriter) Flush() error {
	o := tw.buf.String()
	go func() {
		tw.outCh <- o
	}()
	tw.buf.Reset()
	return nil
}

func (tw *testWriter) waitOutput(cnt int) string {
	var o string
	for i := 0; i < cnt; i++ {
		o = <-tw.outCh
	}
	return o
}

type testElement struct {
	updateCh     chan struct{}
	current      string
	terminalOnly bool
}

func newTestElement(terminalOnly bool) *testElement {
	return &testElement{
		updateCh:     make(chan struct{}),
		terminalOnly: terminalOnly,
	}
}

func (te *testElement) Stop() {}

func (te *testElement) UpdateChan() <-chan struct{} {
	return te.updateCh
}

func (te *testElement) Current() string {
	return te.current
}

func (te *testElement) TerminalOnly() bool {
	return te.terminalOnly
}

func (te *testElement) setCurrent(v string) {
	te.current = v
	te.updateCh <- struct{}{}
}

func TestProgress(t *testing.T) {
	tw := newTestWriter()
	c := NewCounter(100)
	te := newTestElement(false)
	p := new("prefix", LogFunc, tw, true, c, te)
	p.Run()

	require.Equal(t, "prefix 0% (0/100)", tw.waitOutput(1))
	c.SetCount(1)
	require.Equal(t, "prefix 1% (1/100)", tw.waitOutput(1))
	te.setCurrent(" test")
	require.Equal(t, "prefix 1% (1/100) test", tw.waitOutput(1))
	c.SetCount(100)
	te.setCurrent(" test 2")
	require.Equal(t, "prefix 100% (100/100) test 2", tw.waitOutput(2))

	p.Done()
	require.Equal(t, "prefix 100% (100/100) test 2, done.", tw.waitOutput(1))
}

func TestProgressAbort(t *testing.T) {
	tw := newTestWriter()
	te := newTestElement(false)
	p := new("prefix", LogFunc, tw, true, te)
	p.Run()

	require.Equal(t, "prefix", tw.waitOutput(1))
	te.setCurrent(" test")
	require.Equal(t, "prefix test", tw.waitOutput(1))

	p.Abort()
	require.Equal(t, "prefix test", tw.waitOutput(1))
}

func TestWriterInTests(t *testing.T) {
	p := New("", LogFunc)
	require.IsType(t, &standardWriter{}, p.writer)
}
