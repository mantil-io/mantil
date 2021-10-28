package event

import (
	jsoniter "github.com/json-iterator/go"
)

/*
 gomodifytags -all -add-tags short -add-options short=omitempty -transform=camelcase -w -file event.go
 gomodifytags -all -add-tags bson  -add-options bson=omitempty  -transform=camelcase -w -file event.go
 gomodifytags -all -add-tags json  -add-options json=omitempty  -transform=camelcase -w -file event.go
*/

// event raised after execution of a cli command
type CliCommand struct {
	Timestamp int64    `short:"t,omitempty"`
	Version   string   `short:"v,omitempty"`
	Command   string   `short:"c,omitempty"`
	Args      []string `short:"a,omitempty"`
	Workspace string   `short:"w,omitempty"`
	Project   string   `short:"p,omitempty"`
	Stage     string   `short:"s,omitempty"`
	Events    []Event  `short:"e,omitempty"`
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

// placeholder for all events
// only one attribute is not nil
type Event struct {
	GoBuild *GoBuild `short:"g,omitempty"`
	Deploy  *Deploy  `short:"d,omitempty"`
}

type GoBuild struct {
}

type Deploy struct {
	BuildDuration  int64 `short:"b,omitempty"`
	UploadDuration int64 `short:"u,omitempty"`
	UploadMiB      int64 `short:"m,omitempty"`
	UpdateDuration int64 `short:"d,omitempty"`
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
