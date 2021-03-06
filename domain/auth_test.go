package domain

import (
	"context"
	"encoding/base64"
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
		Role:      User,
	}
	ctx := make(map[string]interface{})
	StoreUserClaims(&c, ctx)
	claimsEncoded := base64.StdEncoding.EncodeToString([]byte("{\"w\":\"workspace\",\"p\":\"project\",\"s\":\"stage\",\"r\":\"runtime\",\"u\":\"username\",\"o\":1}"))
	require.Equal(t, ctx["mantilUserClaims"], claimsEncoded)
}

func TestReadUserClaims(t *testing.T) {
	c := AccessTokenClaims{
		Workspace: "workspace",
		Project:   "project",
		Stage:     "stage",
		Runtime:   "runtime",
		Username:  "username",
		Role:      User,
	}

	publicKey, privateKey, err := token.KeyPair()
	require.NoError(t, err)

	headers := make(map[string]string)

	_, err = ReadAccessToken(headers, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "token not found")

	token, err := token.JWT(privateKey, c, 7*24*time.Hour)
	require.NoError(t, err)

	headers[strings.ToLower(AccessTokenHeader)] = token
	_, err = ReadAccessToken(headers, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid key")

	c2, err := ReadAccessToken(headers, publicKey)
	require.NoError(t, err)
	require.Equal(t, &c, c2)

	headers[strings.ToLower(AccessTokenHeader)] = token
	c2, err = ReadAccessToken(headers, publicKey)
	require.NoError(t, err)
	require.Equal(t, &c, c2)

	headers[AccessTokenHeader] = "foo"
	_, err = ReadAccessToken(headers, publicKey)
	require.Error(t, err)
}

func TestClaimsFromContext(t *testing.T) {
	ctx := context.Background()
	c, err := ClaimsFromContext(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "lambda context not found")
	require.Nil(t, c)

	ac := map[string]interface{}{}
	c, err = claimsFromAuthorizerContext(ac)
	require.Error(t, err)
	require.Contains(t, err.Error(), "claims not found")
	require.Nil(t, c)

	ac = map[string]interface{}{
		ContextUserClaimsKey: base64.StdEncoding.EncodeToString([]byte("{\"w\":\"workspace\",\"p\":\"project\",\"s\":\"stage\",\"r\":\"runtime\",\"u\":\"username\",\"o\":1}")),
	}
	c, err = claimsFromAuthorizerContext(ac)
	require.Nil(t, err)
	require.Equal(t, &AccessTokenClaims{
		Workspace: "workspace",
		Project:   "project",
		Stage:     "stage",
		Runtime:   "runtime",
		Username:  "username",
		Role:      User,
	}, c)
}
