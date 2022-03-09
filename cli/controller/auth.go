package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"

	"github.com/mantil-io/mantil.go/logs"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/secret"
	"github.com/mantil-io/mantil/domain"
	"github.com/nats-io/nats.go"
	"github.com/pkg/browser"
)

const (
	clientID = "db4946aabe86cd6c126e"
)

func AuthToken(n *domain.Node) (string, error) {
	t, err := n.AuthToken()
	var terr *domain.TokenExpiredError
	if errors.As(err, &terr) && n.GithubAuthEnabled() {
		var err error
		t, err = githubAuth(n.Endpoints.Rest)
		if err != nil {
			return "", log.Wrap(err)
		}
		n.UpdateToken(t)
	} else if err != nil {
		return "", log.Wrap(err)
	}
	return t, nil
}

func githubAuth(nodeEndpoint string) (string, error) {
	s, err := createState(nodeEndpoint)
	if err != nil {
		return "", err
	}
	if err := githubLogin(s); err != nil {
		return "", err
	}
	t, err := waitToken(s.Inbox)
	if err != nil {
		return "", err
	}
	return t, nil
}

type state struct {
	Inbox        string `json:"inbox"`
	NodeEndpoint string `json:"node_endpoint"`
}

func createState(nodeEndpoint string) (*state, error) {
	inbox := nats.NewInbox()
	s := state{
		Inbox:        inbox,
		NodeEndpoint: nodeEndpoint,
	}
	return &s, nil
}

func githubLogin(state *state) error {
	u, err := url.Parse("https://github.com/login/oauth/authorize")
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("scope", "read:org")
	buf, err := json.Marshal(state)
	if err != nil {
		return err
	}
	sb64 := base64.StdEncoding.EncodeToString([]byte(buf))
	q.Set("state", sb64)
	u.RawQuery = q.Encode()
	return browser.OpenURL(u.String())
}

func waitToken(inbox string) (string, error) {
	rsp := struct {
		JWT string `json:"jwt"`
	}{}
	lc := logs.ListenerConfig{
		ListenerJWT: secret.LogsListenerCreds,
		Subject:     inbox,
		Rsp:         &rsp,
	}
	l, err := logs.NewLambdaListener(lc)
	if err != nil {
		return "", err
	}
	if err := l.Done(context.Background()); err != nil {
		return "", err
	}
	return rsp.JWT, nil
}
