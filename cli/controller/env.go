package controller

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/workspace"
)

type EnvArgs struct {
	Url   bool
	Stage string
}

func Env(a EnvArgs) (string, error) {
	fs, err := NewStoreWithStage(a.Stage)
	if err != nil {
		return "", log.Wrap(err)
	}
	return printEnv(fs.Stage(a.Stage), a.Url)
}

func printEnv(stage *workspace.Stage, onlyURL bool) (string, error) {
	rest := stage.Endpoints.Rest
	if onlyURL {
		return fmt.Sprintf("%s", rest), nil
	}
	return fmt.Sprintf(`export %s='%s'
export %s='%s'
`, workspace.EnvProjectName, stage.Project().Name,
		workspace.EnvApiURL, rest,
	), nil
}
