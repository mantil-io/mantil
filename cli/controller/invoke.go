package controller

import (
	"github.com/mantil-io/mantil/cli/log"
)

type InvokeArgs struct {
	Path           string
	Data           string
	IncludeHeaders bool
	IncludeLogs    bool
	Stage          string
}

func Invoke(a InvokeArgs) error {
	fs, err := newStoreWithStage(a.Stage)
	if err != nil {
		return log.Wrap(err)
	}
	return InvokeCallback(fs.Stage(a.Stage), a.Path, a.Data, a.IncludeHeaders, a.IncludeLogs)()
}
