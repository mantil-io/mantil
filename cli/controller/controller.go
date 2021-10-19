package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/backend"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/workspace"
)

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
	noProjectFoundMsg := "Mantil project not found"
	fs, err := workspace.NewSingleDeveloperProjectStore()
	if err != nil {
		if err == workspace.ErrProjectNotFound {
			return nil, log.WithUserMessage(err, noProjectFoundMsg)
		}
		return nil, log.Wrap(err)
	}
	project := fs.Project()
	if project == nil {
		return nil, log.WithUserMessage(nil, noProjectFoundMsg)
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
