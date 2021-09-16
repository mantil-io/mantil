package auth

import (
	"crypto/ed25519"
	"crypto/rand"

	"github.com/kataras/jwt"
)

func CreateKeyPair() (string, string, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return string(jwt.Base64Encode(publicKey)), string(jwt.Base64Encode(privateKey)), nil
}
