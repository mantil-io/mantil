package terraform

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	logPrefix       = "TF: "
	outputLogPrefix = "TFO: "
)

var (
	createdRegExp            = regexp.MustCompile(`TF: \w*\.\w*\.(\w*)\.(\w*)\[*\"*([\$\w/]*)"*\]*: Creation complete after (\w*)`)
	createdRegExpSubModule   = regexp.MustCompile(`TF: \w*\.\w*\.\w*\..*\.(\w*[\[\]0-9]?)\.(\w*)\[*\"*([\$\w/]*)"*\]*: Creation complete after (\w*)`)
	destroyedRegExp          = regexp.MustCompile(`TF: \w*\.\w*\.(\w*)\.(\w*)\[*\"*([\$\w/]*)"*\]*: Destruction complete after (\w*)`)
	destroyedRegExpSubModule = regexp.MustCompile(`TF: \w*\.\w*\.\w*\..*\.(\w*)\.(\w*)\[*\"*([\$\w/]*)"*\]*: Destruction complete after (\w*)`)
	modifiedRegExp           = regexp.MustCompile(`TF: \w*\.\w*\.(\w*)\.(\w*)\[*\"*([\$\w/]*)"*\]*: Modifications complete after (\w*)`)
	completeRegExp           = regexp.MustCompile(`TF: Apply complete! Resources: (\w*) added, (\w*) changed, (\w*) destroyed.`)
	planRegExp               = regexp.MustCompile(`TF: Plan: (\w*) to add, (\w*) to change, (\w*) to destroy.`)
	outputRegExp             = regexp.MustCompile(`TFO: (\w*) = "(.*)"`)
)

type ParserState int

const (
	StateInitial ParserState = iota
	StateInitializing
	StatePlanning
	StateCreating
	StateDestroying
	StateDone
)

type Parser struct {
	Outputs map[string]string
	counter *resourceCounter
	state   ParserState
}

// NewLogParser creates terraform log parser.
// Prepares log lines for showing to the user.
// Collects terraform output in Outputs map.
func NewLogParser() *Parser {
	p := &Parser{
		Outputs: make(map[string]string),
	}
	return p
}

// Parse terraform line returns false if this is not log line from terraform.
func (p *Parser) Parse(line string) bool {
	if !(strings.HasPrefix(line, logPrefix) ||
		strings.HasPrefix(line, outputLogPrefix)) {
		return false
	}
	if strings.HasPrefix(line, "TFO: >> terraform output") {
		return true
	}
	matchers := []func(string) bool{
		func(line string) bool {
			match := outputRegExp.FindStringSubmatch(line)
			if len(match) == 3 {
				p.Outputs[match[1]] = match[2]
				return true
			}
			return false
		},
		func(line string) bool {
			if strings.HasPrefix(line, "TF: >> terraform init") && !p.isApplying() {
				p.state = StateInitializing
				return true
			}
			return false
		},
		func(line string) bool {
			if strings.HasPrefix(line, "TF: >> terraform plan") && !p.isApplying() {
				p.state = StatePlanning
				return true
			}
			return false
		},
		func(line string) bool {
			if strings.HasPrefix(line, "TF: >> terraform apply") && !p.isApplying() {
				if strings.Contains(line, "-destroy") {
					p.state = StateDestroying
				} else {
					p.state = StateCreating
				}
				return true
			}
			return false
		},
		func(line string) bool {
			if createdRegExp.MatchString(line) ||
				createdRegExpSubModule.MatchString(line) ||
				destroyedRegExp.MatchString(line) ||
				destroyedRegExpSubModule.MatchString(line) ||
				modifiedRegExp.MatchString(line) {
				p.counter.inc()
				return true
			}
			return false
		},
		func(line string) bool {
			if completeRegExp.MatchString(line) {
				p.counter.done()
				p.state = StateDone
				return true
			}
			return false
		},
		func(line string) bool {
			match := planRegExp.FindStringSubmatch(line)
			if len(match) == 4 {
				if p.counter != nil && p.counter.totalCount > 0 {
					return true
				}
				toCreate, _ := strconv.Atoi(match[1])
				toModify, _ := strconv.Atoi(match[2])
				toDestroy, _ := strconv.Atoi(match[3])
				p.counter = newResourceCounter(toCreate + toModify + toDestroy)
				return true
			}
			return false
		},
		func(line string) bool {
			if strings.HasPrefix(line, "TF: Error") {
				p.state = StateDone
				return true
			}
			return false
		},
	}
	for _, m := range matchers {
		if updated := m(line); updated {
			return true
		}
	}
	return true
}

func (p *Parser) Output() string {
	switch p.state {
	case StateInitial:
		return ""
	case StateInitializing:
		return "\tInitializing"
	case StatePlanning:
		return "\tPlanning changes"
	case StateCreating, StateDestroying:
		return p.createDestroyOutput()
	}
	return ""
}

func (p *Parser) createDestroyOutput() string {
	var stateLabel string
	if p.state == StateCreating {
		stateLabel = "Creating"
	}
	if p.state == StateDestroying {
		stateLabel = "Destroying"
	}
	return fmt.Sprintf("\t%s infrastructure %s", stateLabel, p.counter.current())
}

func (p *Parser) State() ParserState {
	return p.state
}

func (p *Parser) isApplying() bool {
	return p.state == StateCreating || p.state == StateDestroying
}

type resourceCounter struct {
	totalCount   int
	currentCount int
}

func newResourceCounter(total int) *resourceCounter {
	r := &resourceCounter{
		totalCount: total,
	}
	return r
}

func (r *resourceCounter) inc() {
	r.currentCount++
	if r.currentCount > r.totalCount {
		r.currentCount = r.totalCount
	}
}

func (r *resourceCounter) done() {
	r.currentCount = r.totalCount
}

func (r *resourceCounter) current() string {
	c := fmt.Sprintf("%d%% (%d/%d)",
		int(100*float64(r.currentCount)/float64(r.totalCount)),
		r.currentCount,
		r.totalCount,
	)
	c = strings.ReplaceAll(c, "%", "%%")
	return c
}
