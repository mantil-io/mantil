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
	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/terraform"
	"github.com/mantil-io/mantil/workspace"
)

var (
	ErrStageNotSet    = fmt.Errorf("stage not set")
	ErrStageNotExists = fmt.Errorf("stage doesn't exist")
)

type Context struct {
	Workspace *workspace.Workspace
	Account   *workspace.Account
	Project   *workspace.Project
	Stage     *workspace.Stage
	Path      string
}

func ContextWithStage(stageName string) (*Context, error) {
	c, err := NewContext()
	if err != nil {
		return nil, log.Wrap(err)
	}
	s := c.ResolveStage(stageName)
	if s == nil {
		return nil, log.Wrap(ErrStageNotExists)
	}
	if err := c.SetStage(s); err != nil {
		return nil, log.Wrap(err)
	}
	return c, nil
}

func NewContext() (*Context, error) {
	w, err := workspace.Load()
	if err != nil {
		return nil, log.Wrap(err)
	}
	path, err := workspace.FindProjectRoot(".")
	if err != nil {
		return nil, log.Wrap(err)
	}
	p, err := workspace.LoadProject(path)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &Context{
		Workspace: w,
		Project:   p,
		Path:      path,
	}, nil
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
	token, err := c.authToken(method)
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
	token, err := c.authToken("")
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	header.Add(auth.AccessTokenHeader, token)
	l, err := logs.NewNATSListener()
	if err != nil {
		return nil, err
	}
	req.Header.Add(logs.InboxHeaderKey, l.Subject())
	req.Header.Add(logs.StreamingTypeHeaderKey, logs.StreamingTypeNATS)
	tp := terraform.NewLogParser()
	err = l.Listen(context.Background(), func(msg string) error {
		if l, ok := tp.Parse(msg); ok {
			if l != "" {
				ui.Info(l)
			}
			log.Printf(msg)
			return nil
		}
		ui.Backend(msg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return l.Wait, nil
}

func (c *Context) authToken(method string) (string, error) {
	if c.Account == nil {
		return "", ErrStageNotSet
	}
	claims := &auth.AccessTokenClaims{
		Workspace: c.Workspace.Name,
		Project:   c.Project.Name,
		Method:    method,
	}
	if c.Stage != nil {
		claims.Stage = c.Stage.Name
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

func (c *Context) AWSClient() (*aws.AWS, error) {
	req := &dto.SecurityRequest{
		ProjectName: c.Project.Name,
		CliRole:     c.Account.CliRole,
	}
	if c.Stage != nil {
		req.StageName = c.Stage.Name
	}
	resp := &dto.SecurityResponse{}
	if err := c.RuntimeRequest("security", req, resp, false); err != nil {
		return nil, err
	}
	awsClient, err := aws.NewWithCredentials(resp.AccessKeyID, resp.SecretAccessKey, resp.SessionToken, resp.Region)
	if err != nil {
		return nil, err
	}
	return awsClient, nil
}

type InvokeArgs struct {
	Path           string
	Data           string
	IncludeHeaders bool
	IncludeLogs    bool
	Stage          string
}

func Invoke(a InvokeArgs) error {
	ctx, err := ContextWithStage(a.Stage)
	if err != nil {
		return log.Wrap(err)
	}
	return ctx.ProjectRequest(a.Path, a.Data, a.IncludeHeaders, a.IncludeLogs)
}

type EnvArgs struct {
	Url   bool
	Stage string
}

func Env(a EnvArgs) (string, error) {
	ctx, err := ContextWithStage(a.Stage)
	if err != nil {
		return "", log.Wrap(err)
	}
	return ctx.env(a)
}

func (ctx *Context) env(a EnvArgs) (string, error) {
	stageURL, err := ctx.StageRestEndpoint()
	if err != nil {
		return "", log.Wrap(err)
	}
	if a.Url {
		return fmt.Sprintf("%s", stageURL), nil
	}
	return fmt.Sprintf(`export %s='%s'
export %s='%s'
`, workspace.EnvProjectName, ctx.Project.Name,
		workspace.EnvApiURL, stageURL,
	), nil
}
