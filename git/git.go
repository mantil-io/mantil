package git

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const (
	OriginRemote = "origin"
)

type Client struct {
	auth *http.BasicAuth
	repo *git.Repository
}

func New(path, token string) (*Client, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &Client{
		auth: &http.BasicAuth{
			Password: token,
		},
		repo: repo,
	}, nil
}

func (c *Client) Commit(path, msg string) error {
	w, err := c.repo.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(path)
	if err != nil {
		return err
	}
	w.Commit(msg, &git.CommitOptions{})
	if err := c.repo.Push(&git.PushOptions{
		Auth: c.auth,
	}); err != nil {
		return err
	}
	return nil
}

func (c *Client) OriginData() (*OriginData, error) {
	config, err := c.repo.Config()
	if err != nil {
		return nil, err
	}
	origin, ok := config.Remotes[OriginRemote]
	if !ok {
		return nil, fmt.Errorf("origin doesn't exist")
	}
	if len(origin.URLs) == 0 {
		return nil, fmt.Errorf("origin doesn't have URLs")
	}
	url := origin.URLs[0]
	url = strings.TrimPrefix(url, "git@github.com:")
	url = strings.TrimSuffix(url, ".git")
	parts := strings.Split(url, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("origin url in unexpected format")
	}
	return &OriginData{
		User:       parts[0],
		Repository: parts[1],
	}, nil
}

type OriginData struct {
	User       string
	Repository string
}

func (c *Client) Branch() (string, error) {
	h, err := c.repo.Head()
	if err != nil {
		return "", err
	}
	return h.Name().Short(), nil
}
