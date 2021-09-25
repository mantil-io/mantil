package main

import (
	"github.com/mantil-io/mantil/cmd/mantil/cmd"
	"github.com/mantil-io/mantil/internal/cli/commands/setup"
)

var (
	commit        string
	tag           string
	dirty         string
	version       string
	functionsPath string
)

func main() {
	cmd.Execute(setup.Version{
		Commit:        commit,
		Tag:           tag,
		Dirty:         len(dirty) > 0,
		FunctionsPath: functionsPath,
		Version:       version,
	})
}
