package domain

import (
	"encoding/json"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/mantil-io/mantil/kit/gz"
)

/*
 gomodifytags -all -add-tags short -add-options short=omitempty -transform=camelcase -w -file event.go
 gomodifytags -all -add-tags bson  -add-options bson=omitempty  -transform=camelcase -w -file event.go
 gomodifytags -all -add-tags json  -add-options json=omitempty  -transform=camelcase -w -file event.go
*/

// event raised after execution of a cli command
type CliCommand struct {
	Timestamp int64      `short:"t,omitempty" json:"timestamp,omitempty"`
	Duration  int64      `short:"d,omitempty" json:"duration,omitempty"`
	Version   string     `short:"v,omitempty" json:"version,omitempty"`
	Command   string     `short:"c,omitempty" json:"command,omitempty"`
	Args      []string   `short:"a,omitempty" json:"args,omitempty"`
	Workspace string     `short:"w,omitempty" json:"workspace,omitempty"`
	Project   string     `short:"p,omitempty" json:"project,omitempty"`
	Stage     string     `short:"s,omitempty" json:"stage,omitempty"`
	Errors    []CliError `short:"r,omitempty" json:"errors,omitempty"`
	Events    []Event    `short:"e,omitempty" json:"events,omitempty"`
}

type CliError struct {
	Error        string `short:"e,omitempty" json:"error,omitempty"`
	Type         string `short:"t,omitempty" json:"type,omitempty"`
	SourceFile   string `short:"s,omitempty" json:"sourceFile,omitempty"`
	FunctionName string `short:"f,omitempty" json:"functionName,omitempty"`
}

func (c *CliCommand) Marshal() ([]byte, error) {
	buf, err := shortConfig.Marshal(c)
	if err != nil {
		return nil, err
	}
	return gz.Zip(buf)
}

func (c *CliCommand) Unmarshal(buf []byte) error {
	buf, err := gz.Unzip(buf)
	if err != nil {
		return err
	}
	return shortConfig.Unmarshal(buf, c)
}

func (c *CliCommand) Pretty() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

func (c *CliCommand) Add(e Event) {
	e.Timestamp = nowMS()
	c.Events = append(c.Events, e)
}

func (c *CliCommand) AddError(e CliError) {
	c.Errors = append(c.Errors, e)
}

// placeholder for all events
// only one attribute is not nil
type Event struct {
	Timestamp int64    `short:"t,omitempty" json:"timestamp,omitempty"`
	GoBuild   *GoBuild `short:"g,omitempty" json:"goBuild,omitempty"`
	Deploy    *Deploy  `short:"d,omitempty" json:"deploy,omitempty"`
	Signal    *Signal  `short:"s,omitempty" json:"signal,omitempty"`
}

type GoBuild struct {
}

type Deploy struct {
	BuildDuration  int64 `short:"b,omitempty" json:"buildDuration,omitempty"`
	UploadDuration int64 `short:"u,omitempty" json:"uploadDuration,omitempty"`
	UploadMiB      int64 `short:"m,omitempty" json:"uploadMiB,omitempty"`
	UpdateDuration int64 `short:"d,omitempty" json:"updateDuration,omitempty"`
}

type Signal struct {
	Name  string `short:"n,omitempty" json:"name,omitempty"`
	Stack string `short:"s,omitempty" json:"stack,omitempty"`
}

// marshal
var shortConfig = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	TagKey:                 "short",
}.Froze()

func short(o interface{}) ([]byte, error) {
	return shortConfig.Marshal(o)
}

func NewCliCommand(buf []byte) (*CliCommand, error) {
	var cc CliCommand
	if err := cc.Unmarshal(buf); err != nil {
		return nil, err
	}
	return &cc, nil
}

func (c *CliCommand) Start() {
	c.Timestamp = nowMS()
}

func (c *CliCommand) End() {
	c.Duration = nowMS() - c.Timestamp
}

func nowMS() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
