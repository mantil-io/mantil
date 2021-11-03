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
}

func NewDotsProgress(lines <-chan string, initLine string) *DotsProgress {
	dp := &DotsProgress{
		lines:       lines,
		currentLine: initLine,
		done:        make(chan struct{}),
	}
	return dp
}

func (dp *DotsProgress) Run() {
	hideCursor()
	go func() {
		<-dp.done
		showCursor()
	}()
	go dp.printLoop()
}

func (dp *DotsProgress) Stop() {
	if dp.isDone() {
		return
	}
	close(dp.done)
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
			dp.currentLine = line
			dp.print()
		case <-dp.done:
			ticker.Stop()
			dp.print()
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
