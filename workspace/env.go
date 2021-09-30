package workspace

import (
	"fmt"
	"log"
)

func Env(stageName string) (string, *Stage) {
	initPath := "."
	path, err := FindProjectRoot(initPath)
	if err != nil {
		log.Fatal(err)
	}
	project, err := LoadProject(path)
	if err != nil {
		log.Fatal(err)
	}
	stage := project.Stage(stageName)
	var url string
	if stage != nil && stage.Endpoints != nil {
		url = stage.Endpoints.Rest
	}
	return fmt.Sprintf(`export %s='%s'
export %s='%s'
`, EnvProjectName, project.Name,
		EnvApiURL, url,
	), stage
}
