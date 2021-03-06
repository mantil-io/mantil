package controller

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/domain"
)

type EnvArgs struct {
	Url   bool
	Stage string
}

func Env(a EnvArgs) (string, error) {
	_, stage, err := newStoreWithStage(a.Stage)
	if err != nil {
		return "", log.Wrap(err)
	}
	return printEnv(stage, a.Url)
}

func printEnv(stage *domain.Stage, onlyURL bool) (string, error) {
	rest := stage.RestEndpoint()
	if onlyURL {
		return fmt.Sprintf("%s", rest), nil
	}
	return fmt.Sprintf(`export %s='%s'
export %s='%s'
`, domain.EnvProjectName, stage.Project().Name,
		domain.EnvApiURL, rest,
	), nil
}
