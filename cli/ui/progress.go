package ui

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

type DotsProgress struct {
	dotCnt      int
	lines       <-chan string
	currentLine string
	done        chan struct{}
	loopDone    chan struct{}
}

func NewDotsProgress(lines <-chan string, initLine string) *DotsProgress {
	dp := &DotsProgress{
		lines:       lines,
		currentLine: initLine,
		done:        make(chan struct{}),
		loopDone:    make(chan struct{}),
	}
	return dp
}

func (dp *DotsProgress) Run() {
	hideCursor()
	go dp.printLoop()
}

func (dp *DotsProgress) Stop() {
	if dp.isDone() {
		return
	}
	close(dp.done)
	<-dp.loopDone
	showCursor()
}

func (dp *DotsProgress) printLoop() {
	dp.print()
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			dp.dotCnt = (dp.dotCnt + 1) % 4
			dp.print()
		case line := <-dp.lines:
			if line != "" {
				dp.currentLine = line
			}
			dp.print()
		case <-dp.done:
			ticker.Stop()
			dp.print()
			close(dp.loopDone)
			return
		}
	}
}

func (dp *DotsProgress) print() {
	var dots string
	format := "\r%s%s, done."
	if !dp.isDone() {
		dots = strings.Repeat(".", dp.dotCnt)
		format = "\r%s%-4s"
	}
	out := fmt.Sprintf(format, dp.currentLine, dots)
	Title(out)
}

func (dp *DotsProgress) isDone() bool {
	select {
	case <-dp.done:
		return true
	default:
		return false
	}
}

func hideCursor() {
	if runtime.GOOS != "windows" {
		fmt.Print("\033[?25l")
	}
}

func showCursor() {
	if runtime.GOOS != "windows" {
		fmt.Print("\033[?25h")
	}
}
