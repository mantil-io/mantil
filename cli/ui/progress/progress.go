package progress

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

type Progress struct {
	prefix     string
	elements   []Element
	done       chan struct{}
	loopDone   chan struct{}
	writer     *writer
	printFunc  func(w io.Writer, format string, v ...interface{})
	isTerminal bool
	closer     sync.Once
}

func New(prefix string, printFunc func(w io.Writer, format string, v ...interface{}), elements ...Element) *Progress {
	p := &Progress{
		prefix:     prefix,
		done:       make(chan struct{}),
		loopDone:   make(chan struct{}),
		writer:     newWriter(colorable.NewColorableStdout()),
		printFunc:  printFunc,
		isTerminal: isTerminal(),
	}
	var els []Element
	for _, e := range elements {
		if e.TerminalOnly() && !p.isTerminal {
			continue
		}
		els = append(els, e)
	}
	p.elements = els
	return p
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
	p.writer.flush()
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

func isTerminal() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	} else {
		return false
	}
}
