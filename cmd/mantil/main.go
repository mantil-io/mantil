package main

import (
	"github.com/mantil-io/mantil/cmd/mantil/cmd"
)

var (
	commit        string
	tag           string
	dirty         string
	version       string
	functionsPath string
)

func main() {
	cmd.Execute(cmd.Version{
		Commit:        commit,
		Tag:           tag,
		Dirty:         len(dirty) > 0,
		FunctionsPath: functionsPath,
		Version:       version,
	})
}
