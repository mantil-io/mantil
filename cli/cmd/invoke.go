package cmd

import (
	"fmt"
	"github.com/mantil-io/mantil/cli/cmd/project"
)

type invokeCmd struct {
	endpoint       string
	ctx            *project.Context
	data           string
	includeHeaders bool
	includeLogs    bool
}

func (c *invokeCmd) run() error {
	url, err := c.url()
	if err != nil {
		return err
	}
	return c.invoke(url)
}

func (c *invokeCmd) url() (string, error) {
	stageURL := c.ctx.Project.RestEndpoint(c.ctx.Stage.Name)
	if stageURL == "" {
		return "", fmt.Errorf("api URL for the stage does not exist")
	}
	return fmt.Sprintf("%s/%s", stageURL, c.endpoint), nil
}

func (c *invokeCmd) invoke(url string) error {
	return c.ctx.ProjectRequest(url, c.data, c.includeHeaders, c.includeLogs)
}
