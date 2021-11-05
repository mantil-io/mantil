package controller

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
)

type InvokeArgs struct {
	Path           string
	Data           string
	IncludeHeaders bool
	ExcludeLogs    bool
	Stage          string
}

func Invoke(a InvokeArgs) error {
	fs, err := newStoreWithStage(a.Stage)
	if err != nil {
		return log.Wrap(err)
	}
	return stageInvokeCallback(fs.Stage(a.Stage), a.Path, a.Data, a.ExcludeLogs, buildShowResponseHandler(a.IncludeHeaders))()
}

func buildShowResponseHandler(includeHeaders bool) func(httpRsp *http.Response) error {
	return func(httpRsp *http.Response) error {
		if isSuccessfulResponse(httpRsp) {
			ui.Notice(httpRsp.Status)
		} else {
			ui.Errorf(httpRsp.Status)
		}

		if includeHeaders {
			printRspHeaders(httpRsp)
			ui.Info("")
		} else if !isSuccessfulResponse(httpRsp) {
			printApiErrorHeader(httpRsp)
		}

		buf, err := ioutil.ReadAll(httpRsp.Body)
		if err != nil {
			return err
		}
		if string(buf) != "" {
			dst := &bytes.Buffer{}
			if err := json.Indent(dst, buf, "", "   "); err != nil {
				ui.Info(string(buf))
			} else {
				ui.Info(dst.String())
			}
		}
		return nil
	}
}

func isSuccessfulResponse(rsp *http.Response) bool {
	return strings.HasPrefix(rsp.Status, "2")
}

func printRspHeaders(rsp *http.Response) {
	for k, v := range rsp.Header {
		ui.Info("%s: %s", k, strings.Join(v, ","))
	}
}

func printApiErrorHeader(rsp *http.Response) {
	header := "X-Api-Error" // TODO remove magic string, use constant same in all places and libraries
	apiErr := rsp.Header.Get(header)
	if apiErr != "" {
		ui.Info("%s: %s", header, apiErr)
	}
}
