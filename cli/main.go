package main

import (
	"github.com/mantil-io/mantil/cli/cmd"
	"github.com/mantil-io/mantil/cli/commands/setup"
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
