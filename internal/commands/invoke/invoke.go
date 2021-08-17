package invoke

import (
	"github.com/atoz-technology/mantil-cli/internal/commands"
)

func Endpoint(endpoint string, data string, includeHeaders, includeLogs bool) error {
	if err := commands.PrintProjectRequest(endpoint, data, includeHeaders, includeLogs); err != nil {
		return err
	}
	return nil
}
