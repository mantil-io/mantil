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
	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
)

type ProjectContext struct {
	Workspace *WorkspaceConfig
	Account   *AccountConfig
	Project   *config.Project
	Stage     *config.Stage
	Path      string
}

func MustProjectContextWithStage(stageName string) *ProjectContext {
	c, stageExists := MustProjectContext(stageName)
	if !stageExists {
		log.Fatalf("stage %s not found", stageName)
	}
	return c
}

func MustProjectContext(stageName string) (c *ProjectContext, stageExists bool) {
	w, err := LoadWorkspaceConfig()
	if err != nil {
		log.Fatal(err)
	}
	path, err := config.FindProjectRoot(".")
	if err != nil {
		log.Fatal(err)
	}
	p, err := config.LoadProject(path)
	if err != nil {
		log.Fatal(err)
	}
	c = &ProjectContext{
		Workspace: w,
		Project:   p,
		Path:      path,
	}
	s := c.resolveStage(stageName)
	if s == nil {
		return c, false
	}
	a := w.Account(s.Account)
	if a == nil {
		log.Fatalf("account %s not found", s.Account)
	}
	c.Stage = s
	c.Account = a
	return c, true
}

func (c *ProjectContext) resolveStage(stageName string) *config.Stage {
	if stageName != "" {
		return c.Project.Stage(stageName)
	}
	return c.Project.DefaultStage()
}

func (c *ProjectContext) SetStage(s *config.Stage) {
	c.Stage = s
	a := c.Workspace.Account(s.Account)
	if a == nil {
		log.Fatalf("account %s not found", s.Account)
	}
	c.Account = a
}

func (c *ProjectContext) RuntimeRequest(method string, req interface{}, rsp interface{}, logs bool) error {
	token, err := c.authToken()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/%s", c.RestEndpoint(), method)
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

func (c *ProjectContext) RestEndpoint() string {
	return c.Account.Endpoints.Rest
}

func (c *ProjectContext) WsEndpoint() string {
	return fmt.Sprintf("%s/$default", c.Account.Endpoints.Ws)
}

func (c *ProjectContext) logListener(req *http.Request) (func() error, error) {
	token, err := c.authToken()
	if err != nil {
		return nil, err
	}
	header := make(http.Header)
	header.Add(auth.AccessTokenHeader, token)
	l, err := logs.NewListener(c.WsEndpoint(), header)
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

func (c *ProjectContext) authToken() (string, error) {
	claims := &auth.AccessTokenClaims{
		Workspace: c.Workspace.Name,
	}
	return auth.CreateJWT(c.Account.Keys.Private, claims, 7*24*time.Hour)
}

func (c *ProjectContext) ProjectRequest(url string, req string, includeHeaders, includeLogs bool) error {
	buf := []byte(req)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(buf))
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

func (c *ProjectContext) InitialiseAWSSDK() *aws.AWS {
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
	creds := &Credentials{}
	if err := c.RuntimeRequest("security", r, creds, false); err != nil {
		log.Fatal(err)
	}
	awsClient, err := aws.NewWithCredentials(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, creds.Region)
	if err != nil {
		log.Fatal(err)
	}
	return awsClient
}
