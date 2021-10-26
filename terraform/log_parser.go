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
	createdRegExp   = regexp.MustCompile(`TF: \w*\.\w*\.(\w*)\.(\w*)\[*\"*([\$\w/]*)"*\]*: Creation complete after (\w*)`)
	destroyedRegExp = regexp.MustCompile(`TF: \w*\.\w*\.(\w*)\.(\w*)\[*\"*([\$\w/]*)"*\]*: Destruction complete after (\w*)`)
	completeRegExp  = regexp.MustCompile(`TF: Apply complete! Resources: (\w*) added, (\w*) changed, (\w*) destroyed.`)
	planRegExp      = regexp.MustCompile(`TF: Plan: (\w*) to add, (\w*) to change, (\w*) to destroy.`)
	outputRegExp    = regexp.MustCompile(`TFO: (\w*) = "(.*)"`)
)

type Parser struct {
	Outputs map[string]string
}

// NewLogParser creates terraform log parser.
// Prepares log lines for showing to the user.
// Collects terraform output in Outputs map.
func NewLogParser() *Parser {
	return &Parser{
		Outputs: make(map[string]string),
	}
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
				return "Initializing"
			}
			return ""
		},
		func(line string) string {
			if strings.HasPrefix(line, "TF: >> terraform plan") {
				return "Planing changes"
			}
			return ""
		},
		func(line string) string {
			if strings.HasPrefix(line, "TF: >> terraform apply") {
				return "Applying changes"
			}
			return ""
		},
		func(line string) string {
			match := createdRegExp.FindStringSubmatch(line)
			if len(match) == 5 {
				if _, err := strconv.Atoi(match[3]); err == nil {
					match[3] = ""
				}
				return fmt.Sprintf("\tCreated %s %s %s", match[1], match[2], match[3])
			}
			return ""
		},
		func(line string) string {
			match := destroyedRegExp.FindStringSubmatch(line)
			if len(match) == 5 {
				if _, err := strconv.Atoi(match[3]); err == nil {
					match[3] = ""
				}
				return fmt.Sprintf("\tDestroyed %s %s %s", match[1], match[2], match[3])
			}
			return ""
		},
		func(line string) string {
			match := completeRegExp.FindStringSubmatch(line)
			if len(match) == 4 {
				return fmt.Sprintf("\n%s resources added, %s changed, %s destroyed", match[1], match[2], match[3])
			}
			return ""
		},
		func(line string) string {
			match := completeRegExp.FindStringSubmatch(line)
			if len(match) == 4 {
				return fmt.Sprintf("\t%s resources added, %s changed, %s destroyed", match[1], match[2], match[3])
			}
			return ""
		},
		func(line string) string {
			match := planRegExp.FindStringSubmatch(line)
			if len(match) == 4 {
				return fmt.Sprintf("\t%s resources to add, %s to change, %s to destroy", match[1], match[2], match[3])
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
