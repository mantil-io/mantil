package project

import (
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
	q := url.Query()
	// TODO zasto postavljemo prefix na razini projekta koji je use case da to gledam
	logGroupPrefix := project.LogGroupPrefix()
	q.Add(dto.CliRoleQueryParam, account.CliRole)
	if stage != nil {
		logGroupPrefix = stage.LogGroupPrefix()
		q.Add(dto.BucketQueryParam, stage.Public.Bucket)
		q.Add(dto.BucketQueryParam, account.Bucket)
	}
	q.Add(dto.LogGroupsPrefixQueryParam, aws.LambdaLogGroup(logGroupPrefix))
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
	if account == nil {
		return "", ErrStageNotSet
	}
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
