package main

import (
	"github.com/mantil-io/mantil/cli/mantil/cmd"
	"github.com/mantil-io/mantil/cli/mantil/commands/setup"
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
