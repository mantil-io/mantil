package cmd

import (
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"
)

type invokeArgs struct {
	path           string
	data           string
	includeHeaders bool
	includeLogs    bool
	stage          string
}

type invokeCmd struct {
	ctx            *project.Context
	path           string
	data           string
	includeHeaders bool
	includeLogs    bool
}

func newInvoke(a *invokeArgs) (*invokeCmd, error) {
	ctx, err := project.ContextWithStage(a.stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &invokeCmd{
		ctx:            ctx,
		path:           a.path,
		data:           a.data,
		includeHeaders: a.includeHeaders,
		includeLogs:    a.includeLogs,
	}, nil
}

func (c *invokeCmd) run() error {
	return c.invoke()
}

func (c *invokeCmd) invoke() error {
	return c.ctx.ProjectRequest(c.path, c.data, c.includeHeaders, c.includeLogs)
}
