package progress

type Terraform struct {
	parser   *TerraformParser
	progress *Progress
	counter  *Counter
}

func NewTerraform() *Terraform {
	parser := NewTerraformParser()
	p := &Terraform{
		parser: parser,
	}
	return p
}

func (p *Terraform) Parse(line string) {
	oldState := p.parser.State()
	if ok := p.parser.Parse(line); !ok {
		return
	}
	p.updateCounter()
	p.checkState(oldState)
}

func (p *Terraform) checkState(oldState ParserState) {
	newState := p.parser.State()
	if newState == oldState {
		return
	}
	if p.progress != nil {
		p.progress.Done()
		p.counter = nil
	}
	if newState == StateDone {
		return
	}
	p.initProgress()
}

func (p *Terraform) initProgress() {
	var pes []Element
	if p.parser.IsApplying() {
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
