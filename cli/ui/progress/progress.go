package progress

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/fatih/color"
)

type Progress struct {
	prefix     string
	elements   []Element
	done       chan struct{}
	loopDone   chan struct{}
	printFunc  func(format string, v ...interface{})
	isTerminal bool
	closer     sync.Once
}

func New(prefix string, printFunc func(format string, v ...interface{}), elements ...Element) *Progress {
	p := &Progress{
		prefix:     prefix,
		done:       make(chan struct{}),
		loopDone:   make(chan struct{}),
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
	if p.isTerminal {
		out = fmt.Sprintf("\r%s", p.prefix)
	}
	for _, e := range p.elements {
		out += e.Current()
	}
	if p.isDone() {
		out += ", done."
	}
	if p.isTerminal {
		clearLine()
	}
	p.printFunc(out)
	if !p.isTerminal {
		fmt.Println()
	}
}

func (p *Progress) isDone() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}

func clearLine() {
	fmt.Print("\u001b[2K")
}

func LogFunc(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func LogFuncBold() func(string, ...interface{}) {
	return color.New(color.Bold).PrintfFunc()
}

func isTerminal() bool {
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		return true
	} else {
		return false
	}
}

func Trim(cb func(format string, args ...interface{})) func(format string, args ...interface{}) {
	var lastLogLine string
	return func(format string, v ...interface{}) {
		line := fmt.Sprintf(format, v...)
		// if strings.Contains(line, "Planning changes") {
		// 	fmt.Printf("line: `%s`\n", line)
		// 	fmt.Printf("buf: %#v\n", []byte(line))
		// }
		//line = strings.TrimPrefix(line, "\u001b[2K")
		line = strings.Replace(line, "\u001b[2K", "", -1)
		line = strings.Replace(line, "\r", "", -1)
		line = strings.Replace(line, "\n", "", -1)
		line = strings.TrimRight(line, " ")
		line = strings.TrimRight(line, ".")

		//fmt.Printf("line before `%s`\n", line)
		//for i := 0; i < 3; i++ {
		//}
		tsLine := strings.TrimSpace(line)
		if tsLine != "" && tsLine != lastLogLine {
			if cb != nil {
				cb("%s", line)
			} else {
				fmt.Println(line)
			}
			lastLogLine = tsLine
		}
	}
}
