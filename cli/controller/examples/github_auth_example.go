package examples

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/mantil-io/mantil.go/logs"
	"github.com/mantil-io/mantil/cli/secret"
	"github.com/mantil-io/mantil/domain"
	"github.com/nats-io/nats.go"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

const (
	clientID = "db4946aabe86cd6c126e"
)

func NewGithubAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "gh-auth",
		Short:  "authenticate to github",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			node := cmd.Flag("node").Value.String()
			n, err := findNode(node)
			if err != nil {
				return err
			}
			s, err := createState(n)
			if err != nil {
				return err
			}
			if err := githubLogin(s); err != nil {
				return err
			}
			if err := waitToken(s.Inbox); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringP("node", "", domain.DefaultNodeName, "")
	return cmd
}

func findNode(name string) (*domain.Node, error) {
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	if err != nil {
		return nil, err
	}
	w := fs.Workspace()
	if len(w.Nodes) == 0 {
		return nil, fmt.Errorf("no nodes avaiable")
	}
	n := w.FindNode(name)
	if n == nil {
		return nil, fmt.Errorf("node not found")
	}
	return n, nil
}

type state struct {
	Inbox        string `json:"inbox"`
	NodeEndpoint string `json:"node_endpoint"`
}

func createState(n *domain.Node) (*state, error) {
	inbox := nats.NewInbox()
	s := state{
		Inbox:        inbox,
		NodeEndpoint: n.Endpoints.Rest,
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
	buf, err := json.Marshal(state)
	if err != nil {
		return err
	}
	sb64 := base64.StdEncoding.EncodeToString([]byte(buf))
	q.Set("state", sb64)
	u.RawQuery = q.Encode()
	return browser.OpenURL(u.String())
}

func waitToken(inbox string) error {
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
		return err
	}
	if err := l.Done(context.Background()); err != nil {
		return err
	}
	r, _ := json.MarshalIndent(rsp, "", "  ")
	fmt.Println(string(r))
	return nil
}
