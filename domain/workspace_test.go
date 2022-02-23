package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWorkspaceNode(t *testing.T) {
	var w Workspace
	require.Len(t, w.Nodes, 0)

	t.Run("add new node", func(t *testing.T) {
		a, err := w.NewNode("first", "accountID", "region", "bucket", "path", "vTest")
		require.NoError(t, err)
		require.Equal(t, a.AccountID, "accountID")
		require.Equal(t, a.Region, "region")
		require.Equal(t, a.Functions.Bucket, "bucket")
		require.Equal(t, a.Functions.Path, "path")
		require.NotEmpty(t, a.Keys.Public)
		require.NotEmpty(t, a.Keys.Private)
		require.NotEmpty(t, a.ID)
		require.Len(t, w.Nodes, 1)
	})

	t.Run("add existing node", func(t *testing.T) {
		require.True(t, w.nodeExists("first"))
		a, err := w.NewNode("first", "accountID", "region", "bucket", "path", "vTest")
		require.Nil(t, a)
		var ea *NodeExistsError
		require.ErrorAs(t, err, &ea)
		require.Equal(t, "first", ea.Name)
	})
}

func TestWorkspaceNodeNames(t *testing.T) {
	w := Workspace{
		Nodes: []*Node{
			{
				Name: "node1",
			},
			{
				Name: "node2",
			},
		},
	}
	require.Equal(t, []string{"node1", "node2"}, w.NodeNames())
}

func TestWorkspaceRemoveNode(t *testing.T) {
	w := Workspace{
		Nodes: []*Node{
			{
				Name: "node1",
			},
			{
				Name: "node2",
			},
		},
	}

	w.RemoveNode("node1")
	require.Len(t, w.Nodes, 1)
}

func TestNodeResources(t *testing.T) {
	var w Workspace
	a, err := w.NewNode("first", "accountID", "region", "bucket", "path", "vTest")
	require.NoError(t, err)
	require.Len(t, a.Resources(), 8)
}

func TestNodeResourceNamingTemplate(t *testing.T) {
	a := &Node{
		ID: "abcdefg",
	}
	require.Equal(t, "mantil-%s-abcdefg", a.ResourceNamingTemplate())
}

func TestNodeUpgradeVersion(t *testing.T) {
	a := &Node{}
	a.UpgradeVersion("v", "fb", "fp")
	require.Equal(t, a.Version, "v")
	require.Equal(t, a.Functions.Bucket, "fb")
	require.Equal(t, a.Functions.Path, "fp")
}

func TestEventRemoveAwsCredentials(t *testing.T) {
	line := `mantil aws install --aws-access-key-id=AKIAIOSFODNN7EXAMPLE --aws-secret-access-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --aws-region=us-east-1`
	args := strings.Split(line, " ")

	args = RemoveAWSCredentials(args)
	expected := "mantil aws install --aws-access-key-id=*** --aws-secret-access-key=*** --aws-region=us-east-1"
	require.Equal(t, expected, strings.Join(args, " "))

	line = `mantil aws install --aws-access-key-id AKIAIOSFODNN7EXAMPLE --aws-secret-access-key wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY --aws-region us-east-1`
	args = strings.Split(line, " ")
	args = RemoveAWSCredentials(args)
	expected = "mantil aws install --aws-access-key-id *** --aws-secret-access-key *** --aws-region us-east-1"
	require.Equal(t, expected, strings.Join(args, " "))
}
