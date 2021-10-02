package cmd

import (
	"github.com/mantil-io/mantil/cli/cmd/project"
)

type invokeCmd struct {
	path           string
	ctx            *project.Context
	data           string
	includeHeaders bool
	includeLogs    bool
}

func (c *invokeCmd) run() error {
	return c.invoke()
}

func (c *invokeCmd) invoke() error {
	return c.ctx.ProjectRequest(c.path, c.data, c.includeHeaders, c.includeLogs)
}
