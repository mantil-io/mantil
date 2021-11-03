package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"text/template"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/backend"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/dto"
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

func awsClient(account *domain.Account, project *domain.Project, stage *domain.Stage) (*aws.AWS, error) {
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

func authToken(account *domain.Account) (string, error) {
	return account.AuthToken()
}

func Backend(account *domain.Account) (*backend.Backend, error) {
	token, err := authToken(account)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return backend.New(account.Endpoints.Rest, token), nil
}

func InvokeCallback(stage *domain.Stage, path, req string, includeHeaders, includeLogs bool) func() error {
	b := backend.Project(stage.Endpoints.Rest, includeHeaders, includeLogs)
	return func() error {
		return b.Call(path, []byte(req), nil)
	}
}

// ensures that workspace and project exists
func newStore() (*domain.FileStore, error) {
	fs, err := domain.NewSingleDeveloperProjectStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	project := fs.Project()
	if project == nil {
		return nil, domain.ErrProjectNotFound
	}
	return fs, nil
}

// also ensures that project has stage
func newStoreWithStage(stageName string) (*domain.FileStore, error) {
	fs, err := newStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	project := fs.Project()
	if len(project.Stages) == 0 {
		return nil, log.Wrapf("No stages in project")
	}
	if fs.Stage(stageName) == nil {
		return nil, log.Wrapf("Stage %s not found", stageName)
	}
	return fs, nil
}

func renderTemplate(content string, data interface{}) ([]byte, error) {
	fcs := template.FuncMap{
		"join":    strings.Join,
		"toLower": strings.ToLower,
		"title":   strings.Title,
		"first":   first,
	}
	tpl, err := template.New("").Funcs(fcs).Parse(content)
	if err != nil {
		return nil, log.Wrap(err)
	}
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return nil, log.Wrap(err)
	}
	return buf.Bytes(), nil
}

func first(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string(s[0])
}
