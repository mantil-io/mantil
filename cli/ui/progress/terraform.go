package progress

import (
	"fmt"

	"github.com/mantil-io/mantil/node/terraform"
)

type Terraform struct {
	parser   *terraform.Parser
	progress *Progress
	counter  *Counter
	done     chan struct{}
}

func NewTerraform() *Terraform {
	parser := terraform.NewLogParser()
	p := &Terraform{
		parser: parser,
		done:   make(chan struct{}),
	}
	return p
}

func (p *Terraform) Parse(line string) {
	oldState := p.parser.State()
	if ok := p.parser.Parse(line); !ok {
		return
	}
	p.checkState(oldState)
	p.updateCounter()
}

func (p *Terraform) checkState(oldState terraform.ParserState) {
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

func (p *Terraform) initProgress() {
	state := p.parser.State()
	var pes []Element
	if state == terraform.StateCreating || state == terraform.StateDestroying {
		p.counter = NewCounter(p.parser.TotalResourceCount())
		pes = append(pes, p.counter)
	}
	pes = append(pes, NewDots())
	p.progress = New(p.parser.StateLabel(), LogFunc, pes...)
	p.progress.Run()
}

func (p *Terraform) updateCounter() {
	if p.counter == nil {
		return
	}
	p.counter.SetCount(p.parser.CurrentResourceCount())
}

func (p *Terraform) close() {
	if p.isDone() {
		return
	}
	close(p.done)
}

func (p *Terraform) isDone() bool {
	select {
	case <-p.done:
		return true
	default:
		return false
	}
}
