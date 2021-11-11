package progress

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sync"

	"github.com/fatih/color"
	"github.com/mantil-io/mantil/cli/ui/term"
	"github.com/mattn/go-colorable"
)

type flushableWriter interface {
	io.Writer
	Flush() error
}

type standardWriter struct {
	out io.Writer
	buf *bytes.Buffer
}

func newStandardWriter(out io.Writer) *standardWriter {
	return &standardWriter{
		out: out,
		buf: bytes.NewBuffer(nil),
	}
}

func (sw *standardWriter) Write(p []byte) (int, error) {
	return sw.buf.Write(p)
}

func (sw *standardWriter) Flush() error {
	defer sw.buf.Reset()
	if err := sw.buf.WriteByte('\n'); err != nil {
		return err
	}
	if _, err := sw.out.Write(sw.buf.Bytes()); err != nil {
		return err
	}
	return nil
}

type Progress struct {
	prefix     string
	elements   []Element
	done       chan struct{}
	loopDone   chan struct{}
	writer     flushableWriter
	printFunc  func(w io.Writer, format string, v ...interface{})
	isTerminal bool
	closer     sync.Once
}

func New(prefix string, printFunc func(w io.Writer, format string, v ...interface{}), elements ...Element) *Progress {
	p := &Progress{
		prefix:     prefix,
		done:       make(chan struct{}),
		loopDone:   make(chan struct{}),
		writer:     term.NewWriter(colorable.NewColorableStdout()),
		printFunc:  printFunc,
		isTerminal: term.IsTerminal(),
	}
	p.initElements(elements)
	p.initWriter()
	return p
}

func (p *Progress) initElements(elements []Element) {
	var filtered []Element
	for _, e := range elements {
		if e.TerminalOnly() && !p.isTerminal {
			continue
		}
		filtered = append(filtered, e)
	}
	p.elements = filtered
}

func (p *Progress) initWriter() {
	out := colorable.NewColorableStdout()
	if p.isTerminal {
		p.writer = term.NewWriter(out)
	} else {
		p.writer = newStandardWriter(out)
	}
}

func (p *Progress) Run() {
	go p.printLoop()
}

func (p *Progress) Stop() {
	p.closer.Do(func() {
		close(p.done)
	})
	<-p.loopDone
}

func (p *Progress) printLoop() {
	p.print()
	var cases []reflect.SelectCase
	for _, e := range p.elements {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(e.UpdateChan()),
		})
	}
	cases = append(cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(p.done),
	})
	for {
		idx, _, _ := reflect.Select(cases)
		if idx == len(cases)-1 {
			for _, e := range p.elements {
				e.Stop()
			}
			if p.isTerminal {
				p.print()
			}
			close(p.loopDone)
			return
		}
		p.print()
	}
}

func (p *Progress) print() {
	out := p.prefix
	for _, e := range p.elements {
		out += e.Current()
	}
	if p.isDone() {
		out += ", done."
	}
	p.printFunc(p.writer, out)
	p.writer.Flush()
}

func (p *Progress) isDone() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}

func LogFunc(w io.Writer, format string, v ...interface{}) {
	fmt.Fprintf(w, format, v...)
}

func LogFuncBold() func(io.Writer, string, ...interface{}) {
	c := color.New(color.Bold)
	return func(w io.Writer, format string, v ...interface{}) {
		c.Fprintf(w, format, v...)
	}
}
