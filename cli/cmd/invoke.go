package cmd

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/commands"
	"github.com/mantil-io/mantil/config"
)

type invokeCmd struct {
	endpoint       string
	project        *config.Project
	stageName      string
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
	stageURL := c.project.RestEndpoint(c.stageName)
	if stageURL == "" {
		return "", fmt.Errorf("api URL for the stage does not exist")
	}
	return fmt.Sprintf("%s/%s", stageURL, c.endpoint), nil
}

func (c *invokeCmd) invoke(url string) error {
	return commands.PrintProjectRequest(url, c.data, c.includeHeaders, c.includeLogs)
}
