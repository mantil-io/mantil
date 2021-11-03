package domain

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/mantil-io/mantil/kit/token"
)

const (
	AccessTokenHeader    = "Authorization"
	EnvPublicKey         = "MANTIL_PUBLIC_KEY"
	ContextUserClaimsKey = "mantilUserClaims"
)

type AccessTokenClaims struct {
	Workspace  string `json:"w,omitempty"`
	Project    string `json:"p,omitempty"`
	Deployment string `json:"d,omitempty"`
	Runtime    string `json:"r,omitempty"`
	Username   string `json:"u,omitempty"`
	Role       Role   `json:"o,omitempty"`
}

type Role int

const (
	Owner Role = iota
	Maintainer
	Member
)

func ReadAccessToken(headers map[string]string) (*AccessTokenClaims, error) {
	if at, ok := headers[AccessTokenHeader]; ok {
		return decodeAccessToken(at)
	}
	if at, ok := headers[strings.ToLower(AccessTokenHeader)]; ok {
		return decodeAccessToken(at)
	}
	return nil, fmt.Errorf("access token not found in %s header", AccessTokenHeader)
}

func decodeAccessToken(at string) (*AccessTokenClaims, error) {
	key, ok := os.LookupEnv(EnvPublicKey)
	if !ok {
		return nil, fmt.Errorf("key not found in environment variable %s", EnvPublicKey)
	}
	var claims AccessTokenClaims
	if err := token.Decode(at, key, &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

func StoreUserClaims(claims *AccessTokenClaims, context map[string]interface{}) {
	v := reflect.ValueOf(claims).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		context["mantil"+t.Field(i).Name] = v.Field(i).Interface()
	}

	buf, _ := json.Marshal(claims)
	context[ContextUserClaimsKey] = string(buf)
}
