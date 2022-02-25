package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-github/v42/github"
	"github.com/mantil-io/mantil.go/logs"
	"github.com/mantil-io/mantil/cli/secret"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/aws"
	"github.com/mantil-io/mantil/kit/token"
	"golang.org/x/oauth2"
)

type JWTRequest struct {
	Inbox       string `json:"inbox"`
	GithubToken string `json:"github_token"`
}

func (a *Auth) JWT(ctx context.Context, req *JWTRequest) error {
	if err := a.initJWT(req); err != nil {
		return err
	}
	jwt, err := a.generateJWT()
	if err != nil {
		return err
	}
	if err := a.publishJWT(jwt); err != nil {
		return err
	}
	return nil
}

func (a *Auth) initJWT(req *JWTRequest) error {
	a.JWTRequest = req
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: req.GithubToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	a.ghClient = github.NewClient(tc)
	awsClient, err := aws.New()
	if err != nil {
		return err
	}
	param := func(key string) (string, error) {
		path, err := domain.SSMParameterPath(key)
		if err != nil {
			return "", err
		}
		return awsClient.GetSSMParameter(path)
	}
	a.privateKey, err = param(domain.SSMPrivateKey)
	if err != nil {
		return err
	}
	a.githubUsername, _ = param(domain.SSMGithubUserKey)
	a.githubOrg, _ = param(domain.SSMGithubOrgKey)
	return nil
}

func (a *Auth) generateJWT() (string, error) {
	ghUser, _, err := a.ghClient.Users.Get(context.Background(), "")
	if err != nil {
		return "", err
	}
	role, err := a.userRole(ghUser)
	if err != nil {
		return "", err
	}
	switch role {
	case domain.Owner:
		return a.ownerToken(*ghUser.Login)
	case domain.Member:
		// check if user is allowed to access the node
		user, err := a.findUser(*ghUser.Login)
		if err != nil {
			return "", err
		}
		projects, err := a.findProjects()
		if err != nil {
			return "", err
		}
		var repos []string
		for _, p := range projects {
			repos = append(repos, p.Repo)
		}
		return a.memberToken(user.Name, repos)
	default:
		return "", fmt.Errorf("unsupported role")
	}
}

func (a *Auth) userRole(ghUser *github.User) (domain.Role, error) {
	if a.githubUsername == *ghUser.Login {
		return domain.Owner, nil
	}
	if a.githubOrg != "" {
		ou, _, err := a.ghClient.Organizations.GetOrgMembership(context.Background(), *ghUser.Login, a.githubOrg)
		if err != nil {
			return domain.Member, err
		}
		if *ou.Role == "admin" {
			return domain.Owner, nil
		} else {
			return domain.Member, nil
		}
	}
	return domain.Member, fmt.Errorf("could not resolve user role")
}

func (a *Auth) ownerToken(username string) (string, error) {
	return token.JWT(a.privateKey, &domain.AccessTokenClaims{
		Username: username,
		Role:     domain.Owner,
	}, 7*24*time.Hour)
}

func (a *Auth) memberToken(username string, projects []string) (string, error) {
	return token.JWT(a.privateKey, &domain.AccessTokenClaims{
		Username: username,
		Role:     domain.Member,
	}, 1*time.Hour)
}

func (a *Auth) publishJWT(jwt string) error {
	cc := logs.ConnectConfig{
		PublisherJWT: secret.LogsPublisherCreds,
		Subject:      a.JWTRequest.Inbox,
	}
	p, err := cc.Publisher()
	if err != nil {
		return err
	}
	rsp := struct {
		JWT string `json:"jwt"`
	}{
		JWT: jwt,
	}
	buf, err := json.Marshal(rsp)
	if err != nil {
		return err
	}
	if err := p.Data(buf); err != nil {
		return err
	}
	if err := p.Close(); err != nil {
		return err
	}
	return nil
}
