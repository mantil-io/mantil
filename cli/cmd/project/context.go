package project

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
	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
)

var (
	ErrStageNotSet = fmt.Errorf("stage not set")
)

type Context struct {
	Workspace *workspace.Workspace
	Account   *workspace.Account
	Project   *workspace.Project
	Stage     *workspace.Stage
	Path      string
}

func MustContextWithStage(stageName string) *Context {
	c := MustContext()
	s := c.ResolveStage(stageName)
	if s == nil {
		ui.Fatalf("stage %s not found", stageName)
	}
	if err := c.SetStage(s); err != nil {
		ui.Fatal(err)
	}
	return c
}

func MustContext() *Context {
	w, err := workspace.Load()
	if err != nil {
		ui.Fatal(err)
	}
	path, err := workspace.FindProjectRoot(".")
	if err != nil {
		ui.Fatal(err)
	}
	p, err := workspace.LoadProject(path)
	if err != nil {
		ui.Fatal(err)
	}
	return &Context{
		Workspace: w,
		Project:   p,
		Path:      path,
	}
}

func (c *Context) ResolveStage(stageName string) *workspace.Stage {
	if stageName != "" {
		return c.Project.Stage(stageName)
	}
	return c.Project.DefaultStage()
}

func (c *Context) SetStage(s *workspace.Stage) error {
	c.Stage = s
	a := c.Workspace.Account(s.Account)
	if a == nil {
		return fmt.Errorf("account %s not found", s.Account)
	}
	c.Account = a
	return nil
}

func (c *Context) RuntimeRequest(method string, req interface{}, rsp interface{}, logs bool) error {
	token, err := c.authToken()
	if err != nil {
		return err
	}
	restEndpoint, err := c.RuntimeRestEndpoint()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/%s", restEndpoint, method)
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("could not create runtime request - %v", err)
	}
	httpReq.Header.Add(auth.AccessTokenHeader, token)
	if logs {
		wait, err := c.logListener(httpReq)
		if err != nil {
			return fmt.Errorf("could not initialize log listener - %v", err)
		}
		defer wait()
	}
	httpRsp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error during runtime request - %v", err)
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

func (c *Context) RuntimeRestEndpoint() (string, error) {
	if c.Account == nil {
		return "", ErrStageNotSet
	}
	return c.Account.Endpoints.Rest, nil
}

func (c *Context) RuntimeWsEndpoint() (string, error) {
	if c.Account == nil {
		return "", ErrStageNotSet
	}
	return fmt.Sprintf("%s/$default", c.Account.Endpoints.Ws), nil
}

func (c *Context) logListener(req *http.Request) (func() error, error) {
	token, err := c.authToken()
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	header.Add(auth.AccessTokenHeader, token)
	wsEndpoint, err := c.RuntimeWsEndpoint()
	if err != nil {
		return nil, err
	}
	l, err := logs.NewListener(wsEndpoint, header)
	if err != nil {
		return nil, err
	}
	req.Header.Add(logs.InboxHeaderKey, l.Subject())
	req.Header.Add(logs.StreamingTypeHeaderKey, logs.StreamingTypeWs)
	err = l.Listen(context.Background(), func(msg string) error {
		ui.Backend(msg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return l.Wait, nil
}

func (c *Context) authToken() (string, error) {
	if c.Account == nil {
		return "", ErrStageNotSet
	}
	claims := &auth.AccessTokenClaims{
		Workspace: c.Workspace.Name,
	}
	return auth.CreateJWT(c.Account.Keys.Private, claims, 7*24*time.Hour)
}

func (c *Context) ProjectRequest(path, req string, includeHeaders, includeLogs bool) error {
	url, err := c.StageRestEndpoint()
	if err != nil {
		return err
	}
	buf := []byte(req)
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", url, path), bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	var waitLogs func() error
	if includeLogs {
		waitLogs, err = c.logListener(httpReq)
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
			ui.Error(err)
		}
	}

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

	buf, err = ioutil.ReadAll(httpRsp.Body)
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

func (c *Context) StageRestEndpoint() (string, error) {
	if c.Stage == nil {
		return "", ErrStageNotSet
	}
	return c.Stage.Endpoints.Rest, nil
}

func (c *Context) StageWsEndpoint() (string, error) {
	if c.Stage == nil {
		return "", ErrStageNotSet
	}
	return c.Stage.Endpoints.Ws, nil
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
	header := "X-Api-Error"
	apiErr := rsp.Header.Get(header)
	if apiErr != "" {
		ui.Info("%s: %s", header, apiErr)
	}
}

func (c *Context) MustInitialiseAWSSDK() *aws.AWS {
	awsClient, err := c.InitialiseAWSSDK()
	if err != nil {
		ui.Fatal(err)
	}
	return awsClient
}

func (c *Context) InitialiseAWSSDK() (*aws.AWS, error) {
	type req struct {
		ProjectName string
		StageName   string
	}
	r := &req{
		ProjectName: c.Project.Name,
	}
	if c.Stage != nil {
		r.StageName = c.Stage.Name
	}
	type rsp struct {
		AccessKeyID     string
		SecretAccessKey string
		SessionToken    string
		Region          string
	}
	creds := &rsp{}
	if err := c.RuntimeRequest("security", r, creds, false); err != nil {
		return nil, err
	}
	awsClient, err := aws.NewWithCredentials(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, creds.Region)
	if err != nil {
		return nil, err
	}
	return awsClient, nil
}
