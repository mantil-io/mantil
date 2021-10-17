package project

import (
	"fmt"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/workspace"
)

type InvokeArgs struct {
	Path           string
	Data           string
	IncludeHeaders bool
	IncludeLogs    bool
	Stage          string
}

func Invoke(a InvokeArgs) error {
	fs, err := NewStoreWithStage(a.Stage)
	if err != nil {
		return log.Wrap(err)
	}
	return InvokeCallback(fs.Stage(a.Stage), a.Path, a.Data, a.IncludeHeaders, a.IncludeLogs)()
}

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
