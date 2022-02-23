package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v42/github"
	"github.com/mantil-io/mantil.go/logs"
	"github.com/mantil-io/mantil/cli/secret"
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
	pathPrefix := os.Getenv(ssmPrefixEnv)
	publicKey, err := awsClient.GetSSMParameter(fmt.Sprintf("%s/public_key", pathPrefix))
	if err != nil {
		return err
	}
	a.publicKey = publicKey
	privateKey, err := awsClient.GetSSMParameter(fmt.Sprintf("%s/private_key", pathPrefix))
	if err != nil {
		return err
	}
	a.privateKey = privateKey
	return nil
}

type claims struct {
	User     string   `json:"user"`
	Projects []string `json:"projects"`
}

func (a *Auth) generateJWT() (string, error) {
	ghUser, _, err := a.ghClient.Users.Get(context.Background(), "")
	if err != nil {
		return "", err
	}
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
	jwt, err := token.JWT(a.privateKey, &claims{
		User:     user.Name,
		Projects: repos,
	}, 10*time.Minute)
	if err != nil {
		return "", err
	}
	return jwt, nil
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
