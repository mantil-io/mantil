package ui

import (
	"fmt"

	"github.com/mantil-io/mantil/node/terraform"
)

type TerraformProgress struct {
	parser   *terraform.Parser
	progress *Progress
	counter  *Counter
	done     chan struct{}
}

func NewTerraformProgress() *TerraformProgress {
	parser := terraform.NewLogParser()
	p := &TerraformProgress{
		parser: parser,
		done:   make(chan struct{}),
	}
	return p
}

func (p *TerraformProgress) Parse(line string) {
	oldState := p.parser.State()
	if ok := p.parser.Parse(line); !ok {
		return
	}
	p.checkState(oldState)
	p.updateCounter()
}

func (p *TerraformProgress) checkState(oldState terraform.ParserState) {
	newState := p.parser.State()
	if newState == oldState {
		return
	}
	if p.progress != nil {
		p.progress.Stop()
		p.counter = nil
		fmt.Println()
	}
	if newState == terraform.StateDone {
		p.close()
		return
	}
	p.initProgress()
}

func (p *TerraformProgress) initProgress() {
	state := p.parser.State()
	var pes []ProgressElement
	if state == terraform.StateCreating || state == terraform.StateDestroying {
		p.counter = NewCounter(p.parser.TotalResourceCount())
		pes = append(pes, p.counter)
	}
	pes = append(pes, NewDots())
	p.progress = NewProgress(p.parser.Output(), ProgressLogFunc, pes...)
	p.progress.Run()
}

func (p *TerraformProgress) updateCounter() {
	if p.counter == nil {
		return
	}
	p.counter.SetCount(p.parser.CurrentResourceCount())
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
