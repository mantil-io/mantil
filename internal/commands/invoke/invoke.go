package invoke

import (
	"github.com/atoz-technology/mantil-cli/internal/commands"
)

func Endpoint(endpoint string, data string) error {
	if err := commands.ProjectRequest(endpoint, data); err != nil {
		return err
	}
	return nil
}
