package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/workspace"
)

type envCmd struct {
	ctx *project.Context
	url bool
}

func (c *envCmd) run() error {
	output, err := c.output()
	if err != nil {
		return err
	}
	fmt.Printf("%s", output)
	return nil
}

func (c *envCmd) output() (string, error) {
	stageURL, err := c.ctx.StageRestEndpoint()
	if err != nil {
		return "", err
	}
	if c.url && stageURL != "" {
		return fmt.Sprintf("%s", stageURL), nil
	}
	return workspace.Env(c.ctx.Project.Name, stageURL), nil
}
