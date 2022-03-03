package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/v42/github"
	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil.go/logs"
	"github.com/mantil-io/mantil/cli/secret"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/aws"
	"github.com/mantil-io/mantil/kit/token"
	"github.com/mantil-io/mantil/node/dto"
	"golang.org/x/oauth2"
)

type Auth struct {
	JWTRequest    *JWTRequest
	store         *Store
	ghClient      *github.Client
	natsPublisher *logs.Publisher
	privateKey    string
	node          *domain.Node
}

func NewAuth() *Auth {
	s, err := NewStore()
	if err != nil {
		log.Fatal(err)
	}
	n, err := s.FindConfig()
	if err != nil {
		log.Fatal(err)
	}
	return &Auth{
		store: s,
		node:  n,
	}
}

type JWTRequest struct {
	Inbox       string `json:"inbox"`
	GithubToken string `json:"github_token"`
}

func (a *Auth) JWT(ctx context.Context, req *JWTRequest) error {
	if err := a.initJWT(req); err != nil {
		a.publishError(err)
		return err
	}
	jwt, err := a.generateJWT()
	if err != nil {
		a.publishError(err)
		return err
	}
	if err := a.publishJWT(jwt); err != nil {
		a.publishError(err)
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

	cc := logs.ConnectConfig{
		PublisherJWT: secret.LogsPublisherCreds,
		Subject:      a.JWTRequest.Inbox,
	}
	p, err := cc.Publisher()
	if err != nil {
		return err
	}
	a.natsPublisher = p
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
		return a.memberToken(*ghUser.Login)
	default:
		return "", fmt.Errorf("unsupported role")
	}
}

func (a *Auth) userRole(ghUser *github.User) (domain.Role, error) {
	if a.node.GithubID == *ghUser.Login {
		return domain.Owner, nil
	}
	u, err := a.store.FindUser(*ghUser.Login)
	var nerr *mantil.ErrItemNotFound
	if errors.As(err, &nerr) {
		return -1, domain.ErrNotAuthorized
	}
	if err != nil {
		return -1, err
	}
	return u.Role, nil
}

func (a *Auth) ownerToken(username string) (string, error) {
	return token.JWT(a.privateKey, &domain.AccessTokenClaims{
		Username: username,
		Role:     domain.Owner,
	}, 7*24*time.Hour)
}

func (a *Auth) memberToken(username string) (string, error) {
	return token.JWT(a.privateKey, &domain.AccessTokenClaims{
		Username: username,
		Role:     domain.Member,
	}, 1*time.Hour)
}

func (a *Auth) publishJWT(jwt string) error {
	rsp := struct {
		JWT string `json:"jwt"`
	}{
		JWT: jwt,
	}
	buf, err := json.Marshal(rsp)
	if err != nil {
		return err
	}
	if err := a.natsPublisher.Data(buf); err != nil {
		return err
	}
	if err := a.natsPublisher.Close(); err != nil {
		return err
	}
	return nil
}

func (a *Auth) publishError(e error) {
	if err := a.natsPublisher.Error(e); err != nil {
		log.Println(err)
	}
	if err := a.natsPublisher.Close(); err != nil {
		log.Println(err)
	}
}

func (a *Auth) Login() *dto.LoginResponse {
	return &dto.LoginResponse{
		Node: a.node,
	}
}
