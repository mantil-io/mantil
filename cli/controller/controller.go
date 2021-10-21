package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"text/template"
	"time"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/backend"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/workspace"
)

type ArgumentError struct {
	msg string
}

func (a *ArgumentError) Error() string {
	return a.msg
}

func NewArgumentError(format string, v ...interface{}) *ArgumentError {
	msg := fmt.Sprintf(format, v...)
	return &ArgumentError{msg: msg}
}

func AWSClient(account *workspace.Account, project *workspace.Project, stage *workspace.Stage) (*aws.AWS, error) {
	restEndpoint := account.Endpoints.Rest

	url, err := url.Parse(fmt.Sprintf("%s/security", restEndpoint))
	if err != nil {
		return nil, log.Wrap(err)
	}

	req := dto.SecurityRequest{
		CliRole: account.CliRole,
		Buckets: []string{account.Bucket},
	}
	if stage != nil {
		req.Buckets = append(req.Buckets, stage.Public.Bucket)
		req.LogGroupsPrefix = aws.LambdaLogGroup(stage.LogGroupsPrefix())
	}
	buf, err := json.Marshal(req)
	if err != nil {
		return nil, log.Wrap(err)
	}
	enc := base64.StdEncoding.EncodeToString(buf)

	q := url.Query()
	q.Add(dto.RequestQueryParam, enc)
	url.RawQuery = q.Encode()

	token := func() string {
		token, err := authToken(account)
		if err != nil {
			return ""
		}
		return token
	}

	awsClient, err := aws.NewWithEndpointCredentials(url.String(), account.Region, token)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return awsClient, nil
}

func authToken(account *workspace.Account) (string, error) {
	claims := &auth.AccessTokenClaims{
		Workspace: account.WorkspaceName(),
	}
	return auth.CreateJWT(account.Keys.Private, claims, 7*24*time.Hour)
}

func Backend(account *workspace.Account) (*backend.Backend, error) {
	token, err := authToken(account)
	if err != nil {
		return nil, err
	}
	return backend.New(account.Endpoints.Rest, token), nil
}

func InvokeCallback(stage *workspace.Stage, path, req string, includeHeaders, includeLogs bool) func() error {
	b := backend.Project(stage.Endpoints.Rest, includeHeaders, includeLogs)
	return func() error {
		return b.Call(path, []byte(req), nil)
	}
}

// ensures that workspace and project exists
func NewStore() (*workspace.FileStore, error) {
	fs, err := workspace.NewSingleDeveloperProjectStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	project := fs.Project()
	if project == nil {
		return nil, workspace.ErrProjectNotFound
	}
	return fs, nil
}

// also ensures that project has stage
func NewStoreWithStage(stageName string) (*workspace.FileStore, error) {
	fs, err := NewStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	project := fs.Project()
	if len(project.Stages) == 0 {
		return nil, log.WithUserMessage(err, "No stages in project")
		// TODO: info create it with `mantil stage new`
	}
	if fs.Stage(stageName) == nil {
		return nil, log.WithUserMessage(nil, "Stage %s not found", stageName)
	}
	return fs, nil
}

func renderTemplate(content string, data interface{}) (string, error) {
	tpl, err := template.New("").Parse(setupStackTemplate)
	if err != nil {
		return "", log.Wrap(err)
	}
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return "", log.Wrap(err)
	}
	return buf.String(), nil
}
