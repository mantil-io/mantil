package ui

import (
	"fmt"

	"github.com/mantil-io/mantil/node/terraform"
)

type TerraformProgress struct {
	parser       *terraform.Parser
	linesCh      chan string
	dotsProgress *DotsProgress
	done         chan struct{}
}

func NewTerraformProgress() *TerraformProgress {
	parser := terraform.NewLogParser()
	p := &TerraformProgress{
		parser: parser,
		done:   make(chan struct{}),
	}
	return p
}

func (p *TerraformProgress) Parse(line string) bool {
	oldState := p.parser.State()
	if ok := p.parser.Parse(line); !ok {
		return false
	}
	newState := p.parser.State()
	if newState != oldState {
		if p.dotsProgress != nil {
			p.dotsProgress.Stop()
			close(p.linesCh)
			fmt.Println()
		}
		if newState == terraform.StateDone {
			p.close()
			return false
		}
		p.linesCh = make(chan string)
		p.dotsProgress = NewDotsProgress(p.linesCh, p.parser.Output(), ProgressLogFunc)
		p.dotsProgress.Run()
	} else if p.dotsProgress != nil {
		p.line(p.parser.Output())
	}
	return true
}

func (p *TerraformProgress) line(l string) {
	if p.isDone() || p.linesCh == nil {
		return
	}
	p.linesCh <- l
}

func (p *TerraformProgress) close() {
	if p.isDone() {
		return
	}
	close(p.done)
}

func (p *TerraformProgress) isDone() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}
