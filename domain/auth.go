package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mantil-io/mantil.go"
	"github.com/mantil-io/mantil/kit/token"
)

const (
	AccessTokenHeader    = "Authorization"
	EnvPublicKey         = "MANTIL_PUBLIC_KEY"
	ContextUserClaimsKey = "mantilUserClaims"
)

type AccessTokenClaims struct {
	Workspace string   `json:"w,omitempty"`
	Project   string   `json:"p,omitempty"`
	Projects  []string `json:"j,omitempty"`
	Stage     string   `json:"s,omitempty"`
	Runtime   string   `json:"r,omitempty"`
	Username  string   `json:"u,omitempty"`
	Role      Role     `json:"o,omitempty"`
}

type Role int

const (
	Owner Role = iota
	Maintainer
	Member
)

var (
	ErrNotAuthorized = fmt.Errorf("not authorized")
)

func ReadAccessToken(headers map[string]string, publicKey string) (*AccessTokenClaims, error) {
	if at, ok := headers[AccessTokenHeader]; ok {
		return decodeAccessToken(at, publicKey)
	}
	if at, ok := headers[strings.ToLower(AccessTokenHeader)]; ok {
		return decodeAccessToken(at, publicKey)
	}
	return nil, fmt.Errorf("access token not found in %s header", AccessTokenHeader)
}

func decodeAccessToken(at, pk string) (*AccessTokenClaims, error) {
	var claims AccessTokenClaims
	if err := token.Decode(at, pk, &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

func StoreUserClaims(claims *AccessTokenClaims, context map[string]interface{}) {
	buf, _ := json.Marshal(claims)
	context[ContextUserClaimsKey] = string(buf)
}

func IsOwner(ctx context.Context) (bool, error) {
	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		return false, err
	}
	return claims.Role == Owner, nil
}

func IsAuthorizedForProject(ctx context.Context, project string) (bool, error) {
	claims, err := ClaimsFromContext(ctx)
	if err != nil {
		return false, err
	}
	switch claims.Role {
	case Owner:
		return true, nil
	case Member:
		for _, p := range claims.Projects {
			if project == p {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, nil
	}
}

func ClaimsFromContext(ctx context.Context) (*AccessTokenClaims, error) {
	lctx, ok := mantil.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("lambda context not found")
	}
	a := lctx.Authorizer()
	c, ok := a[ContextUserClaimsKey]
	if !ok {
		return nil, fmt.Errorf("claims not found in context")
	}
	buf, ok := c.(string)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}
	var claims AccessTokenClaims
	if err := json.Unmarshal([]byte(buf), &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}
