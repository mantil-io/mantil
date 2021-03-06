package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/kataras/jwt"
)

func KeyPair() (string, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return string(jwt.Base64Encode(publicKey)), string(jwt.Base64Encode(privateKey)), nil
}

func JWT(privateKey string, claims interface{}, maxAge time.Duration) (string, error) {
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

func Verify(token string, publicKey string, claims interface{}) error {
	key, err := jwt.Base64Decode([]byte(publicKey))
	if err != nil {
		return fmt.Errorf("failed to decode key %w", err)
	}

	verifiedToken, err := jwt.Verify(jwt.EdDSA, ed25519.PublicKey(key), []byte(token))
	if err != nil {
		return fmt.Errorf("token verify failed: %w", err)
	}

	return verifiedToken.Claims(&claims)
}

func ExpiresIn(token string) (time.Duration, error) {
	var claims jwt.Claims
	if err := Decode(token, &claims); err != nil {
		return -1, err
	}
	return claims.Timeleft(), nil
}

func Decode(token string, claims interface{}) error {
	unverifiedToken, err := jwt.Decode([]byte(token))
	if err != nil {
		return err
	}
	if err := unverifiedToken.Claims(claims); err != nil {
		return err
	}
	return nil
}
