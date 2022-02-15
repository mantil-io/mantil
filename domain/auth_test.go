package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/mantil-io/mantil/kit/token"
	"github.com/stretchr/testify/require"
)

func TestStoreReadUserClaims(t *testing.T) {
	c := AccessTokenClaims{
		Workspace: "workspace",
		Project:   "project",
		Stage:     "stage",
		Runtime:   "runtime",
		Username:  "username",
		Role:      Maintainer,
	}
	ctx := make(map[string]interface{})
	StoreUserClaims(&c, ctx)
	require.Len(t, ctx, 7)
	require.Equal(t, c.Workspace, ctx["mantilWorkspace"])
	require.Equal(t, c.Project, ctx["mantilProject"])
	require.Equal(t, c.Stage, ctx["mantilStage"])
	require.Equal(t, c.Runtime, ctx["mantilRuntime"])
	require.Equal(t, c.Username, ctx["mantilUsername"])
	require.Equal(t, c.Role, ctx["mantilRole"])
	claimsMarshaled := "{\"w\":\"workspace\",\"p\":\"project\",\"s\":\"stage\",\"r\":\"runtime\",\"u\":\"username\",\"o\":1}"
	require.Equal(t, ctx["mantilUserClaims"], claimsMarshaled)
}

func TestReadUserClaims(t *testing.T) {
	c := AccessTokenClaims{
		Workspace: "workspace",
		Project:   "project",
		Stage:     "stage",
		Runtime:   "runtime",
		Username:  "username",
		Role:      Maintainer,
	}

	publicKey, privateKey, err := token.KeyPair()
	headers := make(map[string]string)

	_, err = ReadAccessToken(headers)
	require.Error(t, err)
	require.Contains(t, err.Error(), "token not found")

	token, err := token.JWT(privateKey, c, 7*24*time.Hour)
	require.NoError(t, err)

	headers[strings.ToLower(AccessTokenHeader)] = token
	_, err = ReadAccessToken(headers)
	require.Error(t, err)
	require.Contains(t, err.Error(), "key not found")

	// happy path
	t.Setenv(EnvPublicKey, publicKey)
	c2, err := ReadAccessToken(headers)
	require.NoError(t, err)
	require.Equal(t, &c, c2)

	headers[strings.ToLower(AccessTokenHeader)] = token
	c2, err = ReadAccessToken(headers)
	require.NoError(t, err)
	require.Equal(t, &c, c2)

	headers[AccessTokenHeader] = "foo"
	_, err = ReadAccessToken(headers)
	require.Error(t, err)
}
