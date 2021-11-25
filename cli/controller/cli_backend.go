package controller

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/mantil-io/mantil/backend/dto"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/domain/signup"
)

const cliBackendURL = "https://ytg5gfkg5k.execute-api.eu-central-1.amazonaws.com"

var backend cliBackend

type cliBackend struct{}

func (b cliBackend) Signup() cliBackendSignup {
	return cliBackendSignup{
		endpoint: apiEndpoint{url: cliBackendURL + "/signup"},
	}
}

func (b cliBackend) Report() cliBackendReport {
	return cliBackendReport{
		endpoint: apiEndpoint{url: cliBackendURL + "/report"},
	}
}

type cliBackendSignup struct {
	endpoint apiEndpoint
}

func (s cliBackendSignup) Activate(req signup.ActivateRequest) (string, error) {
	var jwt string
	return jwt, s.endpoint.Call("activate", req, &jwt)
}

func (s cliBackendSignup) Register(req signup.RegisterRequest) error {
	return s.endpoint.Call("register", req, nil)
}

type cliBackendReport struct {
	endpoint apiEndpoint
}

func (r cliBackendReport) URL(req dto.UploadURLRequest) (dto.UploadURLResponse, error) {
	var rsp dto.UploadURLResponse
	return rsp, r.endpoint.Call("url", req, &rsp)
}

func (r cliBackendReport) Confirm(req dto.ConfirmRequest) error {
	return r.endpoint.Call("confirm", req, nil)
}

type apiEndpoint struct {
	url string
}

func (a apiEndpoint) Call(method string, req, rsp interface{}) error {
	buf, _ := json.Marshal(req)
	url := a.url
	url = a.url + "/" + method
	httpRsp, err := http.Post(url, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return log.Wrap(err)
	}
	if err != nil {
		return log.Wrap(err)
	}
	defer httpRsp.Body.Close()
	if httpRsp.StatusCode == http.StatusNoContent {
		return nil
	}
	if httpRsp.StatusCode != http.StatusOK {
		if apiErr := httpRsp.Header.Get("X-Api-Error"); apiErr != "" {
			return log.Wrapf(apiErr)
		}
		return log.Wrapf("request failed with status code %d", httpRsp.StatusCode)
	}
	if rsp != nil {
		buf, err := ioutil.ReadAll(httpRsp.Body)
		if err != nil {
			return log.Wrap(err)
		}

		switch v := rsp.(type) {
		case []byte:
			rsp = buf
		case *string:
			*v = string(buf)
		default:
			if err := json.Unmarshal(buf, rsp); err != nil {
				return log.Wrap(err)
			}
		}
	}
	return nil
}
