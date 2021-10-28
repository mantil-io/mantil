package event

import (
	"encoding/json"

	jsoniter "github.com/json-iterator/go"
)

/*
 gomodifytags -all -add-tags short -add-options short=omitempty -transform=camelcase -w -file event.go
 gomodifytags -all -add-tags bson  -add-options bson=omitempty  -transform=camelcase -w -file event.go
 gomodifytags -all -add-tags json  -add-options json=omitempty  -transform=camelcase -w -file event.go
*/

// event raised after execution of a cli command
type CliCommand struct {
	Timestamp int64    `short:"t,omitempty" json:"timestamp,omitempty"`
	Version   string   `short:"v,omitempty" json:"version,omitempty"`
	Command   string   `short:"c,omitempty" json:"command,omitempty"`
	Args      []string `short:"a,omitempty" json:"args,omitempty"`
	Workspace string   `short:"w,omitempty" json:"workspace,omitempty"`
	Project   string   `short:"p,omitempty" json:"project,omitempty"`
	Stage     string   `short:"s,omitempty" json:"stage,omitempty"`
	Events    []Event  `short:"e,omitempty" json:"events,omitempty"`
}

func (c *CliCommand) Marshal() ([]byte, error) {
	buf, err := shortConfig.Marshal(c)
	if err != nil {
		return nil, err
	}
	return Gzip(buf)
}

func (c *CliCommand) Unmarshal(buf []byte) error {
	buf, err := Gunzip(buf)
	if err != nil {
		return err
	}
	return shortConfig.Unmarshal(buf, c)
}

func (c *CliCommand) Pretty() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// placeholder for all events
// only one attribute is not nil
type Event struct {
	GoBuild *GoBuild `short:"g,omitempty" json:"goBuild,omitempty"`
	Deploy  *Deploy  `short:"d,omitempty" json:"deploy,omitempty"`
}

type GoBuild struct {
}

type Deploy struct {
	BuildDuration  int64 `short:"b,omitempty" json:"buildDuration,omitempty"`
	UploadDuration int64 `short:"u,omitempty" json:"uploadDuration,omitempty"`
	UploadMiB      int64 `short:"m,omitempty" json:"uploadMiB,omitempty"`
	UpdateDuration int64 `short:"d,omitempty" json:"updateDuration,omitempty"`
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
