package git

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/google/go-github/v37/github"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type GithubClient struct {
	*github.Client
	token string
	org   string
}

func NewGithubClient(org string) (*GithubClient, error) {
	t, err := token()
	if err != nil {
		return nil, fmt.Errorf("could not find GitHub access token")
	}
	c := github.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: t}),
		),
	)
	return &GithubClient{c, t, org}, nil
}

func token() (string, error) {
	t, _ := os.LookupEnv("GITHUB_TOKEN")
	if t != "" {
		return t, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	_, err = exec.LookPath("gh")
	if err != nil {
		return "", err
	}

	tokenFromGhConfig := func() (string, error) {
		cfgFile, err := ioutil.ReadFile(fmt.Sprintf("%s/.config/gh/hosts.yml", home))
		if err != nil {
			return "", err
		}
		type ghCfg struct {
			GitHub struct {
				Token string `yaml:"oauth_token"`
			} `yaml:"github.com"`
		}
		c := &ghCfg{}
		err = yaml.Unmarshal(cfgFile, c)
		if err != nil {
			return "", err
		}
		return c.GitHub.Token, nil
	}

	t, err = tokenFromGhConfig()
	if err != nil || t == "" {
		c := exec.Command("gh", "auth", "login")
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		if err != nil {
			return "", err
		}
		t, err = tokenFromGhConfig()
		if err != nil {
			return "", err
		}
	}
	return t, nil
}

func (c *GithubClient) CreateRepo(name, org string, private bool) (*github.Repository, error) {
	r, _, err := c.Repositories.Create(context.Background(), org, &github.Repository{
		Name:    &name,
		Private: &private,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *GithubClient) DeleteRepo(name string) error {
	u, _, err := c.Users.Get(context.Background(), c.org)
	if err != nil {
		return err
	}
	_, err = c.Repositories.Delete(context.Background(), *u.Login, name)
	return err
}

func (c *GithubClient) AddSecrets(repo string, token string) error {
	if err := c.AddSecret(repo, "MANTIL_TOKEN", token); err != nil {
		return err
	}
	return nil
}

func (c *GithubClient) AddSecret(repo, key, value string) error {
	u, _, err := c.Users.Get(context.Background(), c.org)
	if err != nil {
		return err
	}
	owner := *u.Login
	publicKey, _, err := c.Actions.GetRepoPublicKey(context.Background(), owner, repo)
	if err != nil {
		return err
	}
	encryptedSecret, err := c.encryptSecretWithPublicKey(publicKey, key, value)
	if err != nil {
		return err
	}
	if _, err := c.Actions.CreateOrUpdateRepoSecret(context.Background(), owner, repo, encryptedSecret); err != nil {
		return fmt.Errorf("Actions.CreateOrUpdateRepoSecret returned error: %v", err)
	}
	return nil
}

func (c *GithubClient) encryptSecretWithPublicKey(publicKey *github.PublicKey, secretName string, secretValue string) (*github.EncryptedSecret, error) {
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

func (c *GithubClient) createGithubRepo(name, org string, isPrivate bool) (*github.Repository, error) {
	return c.CreateRepo(name, org, isPrivate)
}

func (c *GithubClient) addGithubWorkflow(projectPath string) error {
	destFolder := fmt.Sprintf("%s/.github/workflows", projectPath)
	err := os.MkdirAll(destFolder, os.ModePerm)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/deploy-workflow.yml", destFolder), []byte(DeployWorkflow), 0644)
	if err != nil {
		return err
	}
	return nil
}
