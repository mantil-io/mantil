package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/kataras/jwt"
)

const (
	AccessTokenHeader    = "Authorization"
	EnvPublicKey         = "MANTIL_PUBLIC_KEY"
	ContextUserClaimsKey = "mantilUserClaims"
)

type AccessTokenClaims struct {
	Workspace string `json:"w,omitempty"`
	Project   string `json:"p,omitempty"`
	Stage     string `json:"s,omitempty"`
	Runtime   string `json:"r,omitempty"`
	Method    string `json:"m,omitempty"`
	Username  string `json:"u,omitempty"`
	Role      Role   `json:"o,omitempty"`
}

type Role int

const (
	Owner Role = iota
	Maintainer
	Member
)

func CreateKeyPair() (string, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return string(jwt.Base64Encode(publicKey)), string(jwt.Base64Encode(privateKey)), nil
}

func CreateJWT(privateKey string, claims interface{}, maxAge time.Duration) (string, error) {
	key, err := jwt.Base64Decode([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to decode key %w", err)
	}
	buf, err := jwt.Sign(jwt.EdDSA, ed25519.PrivateKey(key), claims, jwt.MaxAge(maxAge))
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func ReadAccessToken(req *events.APIGatewayCustomAuthorizerRequestTypeRequest) (*AccessTokenClaims, error) {
	if token, ok := req.Headers[AccessTokenHeader]; ok {
		return DecodeAccessToken(token)
	}
	if token, ok := req.Headers[strings.ToLower(AccessTokenHeader)]; ok {
		return DecodeAccessToken(token)
	}
	return nil, fmt.Errorf("access token not found in %s header", AccessTokenHeader)
}

func Decode(token string, claims interface{}) error {
	key, ok := os.LookupEnv(EnvPublicKey)
	if !ok {
		return fmt.Errorf("key not found in environment variable %s", EnvPublicKey)
	}
	return decode(token, key, claims)
}

func decode(token string, publicKey string, claims interface{}) error {
	key, err := jwt.Base64Decode([]byte(publicKey))
	if err != nil {
		return fmt.Errorf("failed to decode key %w", err)
	}

	verifiedToken, err := jwt.Verify(jwt.EdDSA, ed25519.PublicKey(key), []byte(token))
	if err != nil {
		return fmt.Errorf("token %s verify with public key %s failed %w", token, publicKey, err)
	}

	return verifiedToken.Claims(&claims)
}

func DecodeAccessToken(token string) (*AccessTokenClaims, error) {
	var claims AccessTokenClaims
	if err := Decode(token, &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}

func StoreUserClaims(claims *AccessTokenClaims, rsp *events.APIGatewayCustomAuthorizerResponse) {
	if rsp.Context == nil {
		rsp.Context = make(map[string]interface{})
	}

	v := reflect.ValueOf(claims).Elem()
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		rsp.Context["mantil"+t.Field(i).Name] = v.Field(i).Interface()
	}

	buf, _ := json.Marshal(claims)
	rsp.Context[ContextUserClaimsKey] = string(buf)
}
