package terraform

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	completeRegExp           = regexp.MustCompile(`TF: Apply complete! Resources: (\w*) added, (\w*) changed, (\w*) destroyed.`)
	planRegExp               = regexp.MustCompile(`TF: Plan: (\w*) to add, (\w*) to change, (\w*) to destroy.`)
	outputRegExp             = regexp.MustCompile(`TFO: (\w*) = "(.*)"`)
)

const (
	stateInitial = iota
	stateCreating
	stateDestroying
	stateDone
)

type Parser struct {
	Outputs map[string]string
	counter *resourceCounter
	state   int
	out     chan string
}

// NewLogParser creates terraform log parser.
// Prepares log lines for showing to the user.
// Collects terraform output in Outputs map.
func NewLogParser() *Parser {
	p := &Parser{
		Outputs: make(map[string]string),
		out:     make(chan string),
	}
	go p.outputLoop()
	return p
}

// Parse terraform line returns "", false if this is not log line from terraform.
// Returned string is user formatted, for printing in ui.
func (p *Parser) Parse(line string) (string, bool) {
	if !(strings.HasPrefix(line, logPrefix) ||
		strings.HasPrefix(line, outputLogPrefix)) {
		return "", false
	}
	if strings.HasPrefix(line, "TFO: >> terraform output") {
		return "", true
	}
	matchers := []func(string) string{
		func(line string) string {
			match := outputRegExp.FindStringSubmatch(line)
			if len(match) == 3 {
				p.Outputs[match[1]] = match[2]
			}
			return ""
		},
		func(line string) string {
			if strings.HasPrefix(line, "TF: >> terraform init") {
				if p.state == stateInitial {
					p.out <- p.formatOutput("Initializing", false)
				}
			}
			return ""
		},
		func(line string) string {
			if strings.HasPrefix(line, "TF: >> terraform plan") {
				if p.state == stateInitial {
					p.out <- p.formatOutput("Initializing", true)
					p.out <- p.formatOutput("Planning changes", false)
				}
			}
			return ""
		},
		func(line string) string {
			if createdRegExp.MatchString(line) {
				p.counter.inc()
			}
			return ""
		},
		func(line string) string {
			if createdRegExpSubModule.MatchString(line) {
				p.counter.inc()
			}
			return ""
		},
		func(line string) string {
			if destroyedRegExp.MatchString(line) {
				p.counter.inc()
			}
			return ""
		},
		func(line string) string {
			if destroyedRegExpSubModule.MatchString(line) {
				p.counter.inc()
			}
			return ""
		},
		func(line string) string {
			if completeRegExp.MatchString(line) {
				p.counter.done()
				p.state = stateDone
			}
			return ""
		},
		func(line string) string {
			match := planRegExp.FindStringSubmatch(line)
			if len(match) == 4 {
				if p.state == stateInitial {
					p.out <- p.formatOutput("Planning changes", true)
				}
				if p.counter != nil && p.counter.totalCount > 0 {
					return ""
				}
				total, _ := strconv.Atoi(match[1])
				if total == 0 {
					total, _ = strconv.Atoi(match[3])
					p.state = stateDestroying
				} else {
					p.state = stateCreating
				}
				p.counter = newResourceCounter(total)
				return ""
			}
			return ""
		},
	}
	for _, m := range matchers {
		if l := m(line); l != "" {
			return l, true
		}
	}
	return "", true
}

func (p *Parser) outputLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	var stateLabel string
	for range ticker.C {
		createDestroyOutput := func() string {
			if stateLabel == "" {
				if p.state == stateCreating {
					stateLabel = "Creating"
				}
				if p.state == stateDestroying {
					stateLabel = "Destroying"
				}
			}
			return p.formatOutput(
				fmt.Sprintf("%s infrastructure %s", stateLabel, p.counter.current()),
				p.state == stateDone,
			)
		}
		switch p.state {
		case stateCreating, stateDestroying:
			p.out <- createDestroyOutput()
		case stateDone:
			ticker.Stop()
			p.out <- createDestroyOutput()
		}
	}
}

func (p *Parser) formatOutput(o string, done bool) string {
	o = fmt.Sprintf("\r\t%s", o)
	if done {
		o = fmt.Sprintf("%s, done.\n", o)
	}
	return o
}

func (p *Parser) Out() <-chan string {
	return p.out
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
	return fmt.Sprintf("%d%% (%d/%d)",
		int(100*float64(r.currentCount)/float64(r.totalCount)),
		r.currentCount,
		r.totalCount,
	)
}
