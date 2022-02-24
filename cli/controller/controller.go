package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/mantil-io/mantil/cli/controller/invoke"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/aws"
	"github.com/mantil-io/mantil/node/dto"
	"github.com/olekukonko/tablewriter"
)

// package defers
// enable code to be run after controller finishes
var defers []func()

func addDefer(d func()) {
	defers = append(defers, d)
}

// run package defers
func Defer() {
	for _, d := range defers {
		d()
	}
}

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

func awsClient(node *domain.Node, stage *domain.Stage) (*aws.AWS, error) {
	restEndpoint := node.Endpoints.Rest

	url, err := url.Parse(fmt.Sprintf("%s/security", restEndpoint))
	if err != nil {
		return nil, log.Wrap(err)
	}

	req := dto.SecurityRequest{
		CliRole: node.CliRole,
		Buckets: []string{node.Bucket},
	}
	if stage != nil {
		if stage.HasPublic() {
			req.Buckets = append(req.Buckets, stage.Public.Bucket)
		}
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
		token, err := authToken(node)
		if err != nil {
			return ""
		}
		return token
	}

	awsClient, err := aws.NewWithEndpointCredentials(url.String(), node.Region, token)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return awsClient, nil
}

func nodeInvoker(node *domain.Node) (*invoke.HTTPClient, error) {
	token, err := authToken(node)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return invoke.Node(node.Endpoints.Rest, token, ui.NodeLogsSink), nil
}

func stageInvokeCallback(stage *domain.Stage, path, req string, excludeLogs bool, cb func(*http.Response) error) (func() error, error) {
	token, err := stage.AuthToken()
	if err != nil {
		return nil, log.Wrap(err)
	}
	is := invoke.Stage(stage.RestEndpoint(), excludeLogs, cb, token, ui.InvokeLogsSink)
	return func() error {
		return is.Do(path, []byte(req), nil)
	}, nil
}

func newStore() (*domain.FileStore, error) {
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	addDefer(func() { log.SetStage(fs, nil, nil) })
	return fs, err
}

// ensures that workspace and project exists
func newProjectStore() (*domain.FileStore, *domain.Project, error) {
	fs, err := domain.NewSingleDeveloperProjectStore()
	if err != nil {
		return nil, nil, log.Wrap(err)
	}
	project := fs.Project()
	if project == nil {
		return nil, nil, &domain.ProjectNotFoundError{}
	}
	addDefer(func() { log.SetStage(fs, project, nil) })
	return fs, project, nil
}

// also ensures that project has stage
func newStoreWithStage(stageName string) (*domain.FileStore, *domain.Stage, error) {
	fs, project, err := newProjectStore()
	if err != nil {
		return nil, nil, log.Wrap(err)
	}
	if len(project.Stages) == 0 {
		return nil, nil, log.Wrap(&domain.ProjectNoStagesError{})
	}
	stage := fs.Stage(stageName)
	if stage == nil {
		return nil, nil, log.Wrapf("stage %s not found", stageName)
	}
	addDefer(func() { log.SetStage(fs, project, stage) })
	return fs, stage, nil
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

func timerFn() func() int {
	start := time.Now()
	return func() int {
		dur := int(time.Now().Sub(start) / time.Millisecond)
		start = time.Now()
		return dur
	}
}

func ShowTable(header []string, data [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	for _, row := range data {
		table.Append(row)
	}
	table.Render()
}
