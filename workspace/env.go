package workspace

import (
	"fmt"
)

func Env(projectName, apiURL string) string {
	return fmt.Sprintf(`export %s='%s'
export %s='%s'
`, EnvProjectName, projectName,
		EnvApiURL, apiURL,
	)
}
