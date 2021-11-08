package ui

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/color"
)

type ProgressElement interface {
	UpdateChan() <-chan struct{}
	Current() string
	Stop()
}

type Dots struct {
	currentCnt int
	updateCh   chan struct{}
	done       chan struct{}
}

func NewDots() *Dots {
	d := &Dots{
		updateCh: make(chan struct{}),
		done:     make(chan struct{}),
	}
	go d.loop()
	return d
}

func (d *Dots) loop() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			d.currentCnt = (d.currentCnt + 1) % 4
			d.updateCh <- struct{}{}
		case <-d.done:
			ticker.Stop()
			close(d.updateCh)
			return
		}
	}
}

func (d *Dots) Stop() {
	if d.isDone() {
		return
	}
	close(d.done)
}

func (d *Dots) UpdateChan() <-chan struct{} {
	return d.updateCh
}

func (d *Dots) Current() string {
	if d.isDone() {
		return ""
	}
	dots := strings.Repeat(".", d.currentCnt)
	return fmt.Sprintf("%-4s", dots)
}

func (d *Dots) isDone() bool {
	select {
	case <-d.done:
		return true
	default:
		return false
	}
}

type Counter struct {
	total    int
	current  int
	updateCh chan struct{}
}

func NewCounter(total int) *Counter {
	c := &Counter{
		total:    total,
		updateCh: make(chan struct{}),
	}
	return c
}

func (c *Counter) SetCount(value int) {
	c.current = value
	c.updateCh <- struct{}{}
}

func (c *Counter) Current() string {
	cur := fmt.Sprintf(" %d%% (%d/%d)",
		int(100*float64(c.current)/float64(c.total)),
		c.current,
		c.total,
	)
	cur = strings.ReplaceAll(cur, "%", "%%")
	return cur
}

func (c *Counter) UpdateChan() <-chan struct{} {
	return c.updateCh
}

func (c *Counter) Stop() {}

type Progress struct {
	prefix    string
	elements  []ProgressElement
	done      chan struct{}
	loopDone  chan struct{}
	printFunc func(format string, v ...interface{})
}

func NewProgress(prefix string, printFunc func(format string, v ...interface{}), elements ...ProgressElement) *Progress {
	p := &Progress{
		prefix:    prefix,
		elements:  elements,
		done:      make(chan struct{}),
		loopDone:  make(chan struct{}),
		printFunc: printFunc,
	}
	return p
}

func (p *Progress) Run() {
	go p.printLoop()
}

func (p *Progress) Stop() {
	if p.isDone() {
		return
	}
	close(p.done)
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
			p.print()
			close(p.loopDone)
			return
		}
		p.print()
	}
}

func (p *Progress) print() {
	out := fmt.Sprintf("\r%s", p.prefix)
	for _, e := range p.elements {
		out += e.Current()
	}
	if p.isDone() {
		out += ", done."
	}
	clearLine()
	p.printFunc(out)
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

func ProgressLogFunc(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func ProgressLogFuncBold() func(string, ...interface{}) {
	return color.New(color.Bold).PrintfFunc()
}

func TrimProgress(cb func(format string, args ...interface{})) func(format string, args ...interface{}) {
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
