package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/mantil-io/mantil.go/pkg/streaming/logs"
	"github.com/mantil-io/mantil/internal/auth"
	"github.com/mantil-io/mantil/internal/cli/log"
)

func BackendRequest(method string, req interface{}, rsp interface{}, logs bool) error {
	token, err := authToken()
	if err != nil {
		return err
	}
	restURL, err := DefaultRestEndpoint()
	if err != nil {
		return fmt.Errorf("could not get backend url - %v", err)
	}
	url := fmt.Sprintf("%s/%s", restURL, method)
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("could not create backend request - %v", err)
	}
	httpReq.Header.Add(auth.AccessTokenHeader, token)
	if logs {
		wait, err := logListener(httpReq)
		if err != nil {
			return fmt.Errorf("could not initialize log listener - %v", err)
		}
		defer wait()
	}
	httpRsp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error during backend request - %v", err)
	}
	defer httpRsp.Body.Close()
	apiErr := httpRsp.Header.Get("X-Api-Error")
	if apiErr != "" {
		return fmt.Errorf(apiErr)
	}
	if rsp == nil {
		return nil
	}
	buf, err = ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return fmt.Errorf("could not read response - %v", err)
	}
	err = json.Unmarshal(buf, rsp)
	if err != nil {
		return fmt.Errorf("could not unmarshal response - %v", err)
	}
	return nil
}

func authToken() (string, error) {
	w, err := LoadWorkspaceConfig()
	if err != nil {
		return "", err
	}
	claims := &auth.AccessTokenClaims{
		Workspace: w.Name,
	}
	return auth.CreateJWT(w.DefaultAccount().Keys.Private, claims, 7*24*time.Hour)
}

func PrintProjectRequest(url string, req string, includeHeaders, includeLogs bool) error {
	buf := []byte(req)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	var waitLogs func() error
	if includeLogs {
		waitLogs, err = logListener(httpReq)
		if err != nil {
			return err
		}
	}
	httpRsp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpRsp.Body.Close()

	if waitLogs != nil {
		if err := waitLogs(); err != nil {
			log.Error(err)
		}
	}

	if isSuccessfulResponse(httpRsp) {
		log.Notice(httpRsp.Status)
	} else {
		log.Errorf(httpRsp.Status)
	}

	if includeHeaders {
		printRspHeaders(httpRsp)
		log.Info("")
	} else if !isSuccessfulResponse(httpRsp) {
		printApiErrorHeader(httpRsp)
	}

	buf, err = ioutil.ReadAll(httpRsp.Body)
	if err != nil {
		return err
	}
	if string(buf) != "" {
		dst := &bytes.Buffer{}
		if err := json.Indent(dst, buf, "", "   "); err != nil {
			log.Info(string(buf))
		} else {
			log.Info(dst.String())
		}
	}
	return nil
}

func isSuccessfulResponse(rsp *http.Response) bool {
	return strings.HasPrefix(rsp.Status, "2")
}

func printRspHeaders(rsp *http.Response) {
	for k, v := range rsp.Header {
		log.Info("%s: %s", k, strings.Join(v, ","))
	}
}

func printApiErrorHeader(rsp *http.Response) {
	header := "X-Api-Error"
	apiErr := rsp.Header.Get(header)
	if apiErr != "" {
		log.Info("%s: %s", header, apiErr)
	}
}

func logListener(req *http.Request) (func() error, error) {
	wsURL, err := DefaultWsEndpoint()
	if err != nil {
		return nil, err
	}
	token, err := authToken()
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	header.Add(auth.AccessTokenHeader, token)
	l, err := logs.NewListener(wsURL, header)
	if err != nil {
		return nil, err
	}
	req.Header.Add(logs.InboxHeaderKey, l.Subject())
	req.Header.Add(logs.StreamingTypeHeaderKey, logs.StreamingTypeWs)
	err = l.Listen(context.Background(), func(msg string) error {
		log.Backend(msg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return l.Wait, nil
}

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}
