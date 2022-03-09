package domain

import (
	"testing"
	"time"

	"github.com/kataras/jwt"
	"github.com/mantil-io/mantil/kit/token"
	"github.com/stretchr/testify/require"
)

func TestNodeStore(t *testing.T) {
	ns := &NodeStore{}
	_, privateKey, _ := token.KeyPair()

	tk := nodeToken(&Node{
		Name:    "node1",
		Version: "1",
	}, privateKey)
	err := ns.UpsertNodeToken(tk)
	require.NoError(t, err)
	require.Len(t, ns.Nodes, 1)

	n, err := ns.Node("node1")
	require.NoError(t, err)
	require.NotNil(t, n)
	require.Equal(t, "node1", n.Name)
	require.Equal(t, "1", n.Version)

	n, err = ns.FindNode("node1")
	require.NoError(t, err)
	require.NotNil(t, n)
	require.Equal(t, "node1", n.Name)

	n, err = ns.FindNode("")
	require.NoError(t, err)
	require.NotNil(t, n)
	require.Equal(t, "node1", n.Name)

	n, err = ns.FindNode("node2")
	require.Error(t, err)
	require.Nil(t, n)

	tkn := ns.Token("node1")
	require.Equal(t, tk, tkn)

	err = ns.UpsertNodeToken(nodeToken(&Node{
		Name:    "node1",
		Version: "2",
	}, privateKey))
	require.NoError(t, err)

	n, err = ns.Node("node1")
	require.NoError(t, err)
	require.NotNil(t, n)
	require.Equal(t, "node1", n.Name)
	require.Equal(t, "2", n.Version)

	n, err = ns.Node("node2")
	require.Error(t, err)
	require.Nil(t, n)

	ns.RemoveNode("node1")
	n, err = ns.Node("node1")
	require.Error(t, err)
	require.Nil(t, n)
}

func TestAuthToken(t *testing.T) {
	// single developer auth
	publicKey, privateKey, _ := token.KeyPair()
	n := &Node{
		Name: "node-sd",
		Keys: NodeKeys{
			Public:  publicKey,
			Private: privateKey,
		},
		workspace: &Workspace{
			ID: "workspace",
		},
	}
	token, err := n.AuthToken()
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// github auth
	w := &Workspace{}
	n = &Node{
		Name:       "node-gh",
		GithubUser: "gh-user",
		workspace:  w,
	}
	token = nodeToken(n, privateKey)
	err = w.AddNodeToken(token)
	require.NoError(t, err)

	tk, err := n.AuthToken()
	require.NoError(t, err)
	require.Equal(t, token, tk)

	jwt.Clock = func() time.Time {
		return time.Now().Add(2 * time.Hour)
	}
	_, err = n.AuthToken()
	require.Error(t, err)
}

func nodeToken(n *Node, privateKey string) string {
	c := AccessTokenClaims{
		Node: n,
	}
	t, _ := token.JWT(privateKey, &c, time.Hour)
	return t
}
