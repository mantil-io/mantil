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
	q.Add(dto.ProjectNameQueryParam, project.Name)
	q.Add(dto.CliRoleQueryParam, account.CliRole)
	if stage != nil {
		q.Add(dto.StageNameQueryParam, stage.Name)
		q.Add(dto.BucketQueryParam, account.Bucket)
	}
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
