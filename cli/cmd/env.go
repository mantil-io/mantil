package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/workspace"
)

type envCmd struct {
	url       bool
	stageName string
}

func (c *envCmd) run() error {
	env, stage := workspace.Env(c.stageName)
	fmt.Printf("%s", c.output(env, stage))
	return nil
}

func (c *envCmd) output(env string, stage *workspace.Stage) string {
	if c.url && stage != nil && stage.Endpoints != nil {
		return fmt.Sprintf("%s", stage.Endpoints.Rest)
	}
	return env
}
