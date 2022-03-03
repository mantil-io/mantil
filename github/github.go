package github

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/go-github/v42/github"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
)

type Client struct {
	org string
	*github.Client
}

func New(token string, org string) (*Client, error) {
	c := github.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}),
		),
	)
	return &Client{org: org, Client: c}, nil
}

func (c *Client) AddSecret(repo, key, value string) error {
	u, _, err := c.Users.Get(context.Background(), c.org)
	if err != nil {
		return err
	}
	owner := *u.Login
	publicKey, _, err := c.Actions.GetRepoPublicKey(context.Background(), owner, repo)
	if err != nil {
		return err
	}
	encryptedSecret, err := encryptSecretWithPublicKey(publicKey, key, value)
	if err != nil {
		return err
	}
	if _, err := c.Actions.CreateOrUpdateRepoSecret(context.Background(), owner, repo, encryptedSecret); err != nil {
		return err
	}
	return nil
}

func encryptSecretWithPublicKey(publicKey *github.PublicKey, secretName string, secretValue string) (*github.EncryptedSecret, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey.GetKey())
	if err != nil {
		return nil, fmt.Errorf("base64.StdEncoding.DecodeString was unable to decode public key: %v", err)
	}
	var decodedPublicKeyArr [32]byte
	copy(decodedPublicKeyArr[:], decodedPublicKey[:])

	secretBytes := []byte(secretValue)
	out := []byte{}
	encryptedBytes, err := box.SealAnonymous(out, secretBytes, &decodedPublicKeyArr, nil)
	if err != nil {
		return nil, err
	}

	encryptedString := base64.StdEncoding.EncodeToString(encryptedBytes)
	keyID := publicKey.GetKeyID()
	encryptedSecret := &github.EncryptedSecret{
		Name:           secretName,
		KeyID:          keyID,
		EncryptedValue: encryptedString,
	}
	return encryptedSecret, nil
}
