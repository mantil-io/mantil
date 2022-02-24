package auth

import (
	"context"
	"log"

	"github.com/google/go-github/v42/github"
	"github.com/mantil-io/mantil.go"
)

type Auth struct {
	JWTRequest *JWTRequest
	users      *mantil.KV
	projects   *mantil.KV
	ghClient   *github.Client
	publicKey  string
	privateKey string
}

func New() *Auth {
	a := &Auth{}
	if err := a.initStore(); err != nil {
		log.Fatal(err)
	}
	return a
}

type AddUserRequest struct {
	Username string `json:"username"`
}

func (a *Auth) AddUser(ctx context.Context, req *AddUserRequest) error {
	return a.storeUser(req.Username)
}

type AddProjectRequest struct {
	Repo string `json:"repo"`
}

func (a *Auth) AddProject(ctx context.Context, req *AddProjectRequest) error {
	return a.storeProject(req.Repo)
}
