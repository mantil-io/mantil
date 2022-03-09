package domain_test

import (
	"testing"
	"time"

	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/token"
	"github.com/stretchr/testify/require"
)

func TestNodeStore(t *testing.T) {
	ns := &domain.NodeStore{}
	_, privateKey, _ := token.KeyPair()

	tk := nodeToken(&domain.Node{
		Name: "node1",
	}, privateKey)
	err := ns.UpsertNodeToken(tk)
	require.NoError(t, err)
	require.Len(t, ns.Nodes, 1)

	n, err := ns.Node("node1")
	require.NoError(t, err)
	require.NotNil(t, n)
	require.Equal(t, "node1", n.Name)

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

	n, err = ns.Node("node2")
	require.Error(t, err)
	require.Nil(t, n)

	ns.RemoveNode("node1")
	n, err = ns.Node("node1")
	require.Error(t, err)
	require.Nil(t, n)
}

func nodeToken(n *domain.Node, privateKey string) string {
	c := domain.AccessTokenClaims{
		Node: n,
	}
	t, _ := token.JWT(privateKey, &c, time.Hour)
	return t
}
