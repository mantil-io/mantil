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
	modifiedRegExpSubModule  = regexp.MustCompile(`TF: \w*\.\w*\.\w*\..*\.(\w*)\.(\w*)\[*\"*([\$\w/]*)"*\]*: Modifications complete after (\w*)`)
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
	StateUpdating
	StateDestroying
	StateDone
)

type Parser struct {
	Outputs         map[string]string
	counter         *resourceCounter
	state           ParserState
	collectingError bool
	errorMessage    string
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
	if !(strings.Contains(line, logPrefix) ||
		strings.Contains(line, outputLogPrefix)) {
		return false
	}
	if strings.Contains(line, "TFO: >> terraform output") {
		return true
	}
	matchers := []func(string) bool{
		func(line string) bool {
			if p.isError(line) {
				p.collectingError = true
				p.state = StateDone
			}
			if p.collectingError {
				line = strings.TrimPrefix(line, logPrefix)
				p.errorMessage += line + "\n"
				return true
			}
			return false
		},
		func(line string) bool {
			match := outputRegExp.FindStringSubmatch(line)
			if len(match) == 3 {
				p.Outputs[match[1]] = match[2]
				return true
			}
			return false
		},
		func(line string) bool {
			if strings.Contains(line, "TF: >> terraform init") && !p.IsApplying() {
				p.state = StateInitializing
				return true
			}
			return false
		},
		func(line string) bool {
			if strings.Contains(line, "TF: >> terraform plan") && !p.IsApplying() {
				p.state = StatePlanning
				return true
			}
			return false
		},
		func(line string) bool {
			if createdRegExp.MatchString(line) ||
				createdRegExpSubModule.MatchString(line) ||
				destroyedRegExp.MatchString(line) ||
				destroyedRegExpSubModule.MatchString(line) ||
				modifiedRegExp.MatchString(line) ||
				modifiedRegExpSubModule.MatchString(line) {
				p.counter.inc()
				return true
			}
			return false
		},
		func(line string) bool {
			if completeRegExp.MatchString(line) {
				if p.counter != nil {
					p.counter.done()
				}
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
				if toCreate > 0 && toModify == 0 && toDestroy == 0 {
					p.state = StateCreating
				} else if toCreate == 0 && toModify == 0 && toDestroy > 0 {
					p.state = StateDestroying
				} else {
					p.state = StateUpdating
				}
				p.counter = newResourceCounter(toCreate + toModify + toDestroy)
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

func (p *Parser) StateLabel() string {
	switch p.state {
	case StateInitial:
		return ""
	case StateInitializing:
		return "\tInitializing"
	case StatePlanning:
		return "\tPlanning changes"
	case StateCreating:
		return "\tCreating infrastructure"
	case StateUpdating:
		return "\tUpdating infrastructure"
	case StateDestroying:
		return "\tDestroying infrastructure"
	}
	return ""
}

func (p *Parser) State() ParserState {
	return p.state
}

func (p *Parser) Error() error {
	if p.errorMessage == "" {
		return nil
	}
	return fmt.Errorf(p.errorMessage)
}

func (p *Parser) TotalResourceCount() int {
	if p.counter == nil {
		return 0
	}
	return p.counter.totalCount
}

func (p *Parser) CurrentResourceCount() int {
	if p.counter == nil {
		return 0
	}
	return p.counter.currentCount
}

func (p *Parser) IsApplying() bool {
	return p.state == StateCreating || p.state == StateDestroying || p.state == StateUpdating
}

func (p *Parser) isError(line string) bool {
	if strings.Contains(line, "TF: Error") {
		// skip api gateway conflict errors since we are handling them on the backend
		return !strings.Contains(line, "ConflictException: Unable to complete operation due to concurrent modification. Please try again later.")
	}
	return false
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
