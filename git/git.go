package git

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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

// TODO: is this right?
func (c *Client) Branch() (string, error) {
	h, err := c.repo.Head()
	if err != nil {
		return "", err
	}
	return h.Name().Short(), nil
}
