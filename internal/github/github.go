package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atoz-technology/mantil-cli/internal/assets"
	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/go-github/v37/github"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

func token() (string, error) {
	t, ok := os.LookupEnv("GITHUB_TOKEN")
	if ok {
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

type Client struct {
	*github.Client
	token string
	org   string
}

func NewClient(org string) (*Client, error) {
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
	return &Client{c, t, org}, nil
}

func (c *Client) CreateRepo(name, org string, private bool) (*github.Repository, error) {
	r, _, err := c.Repositories.Create(context.Background(), org, &github.Repository{
		Name:    &name,
		Private: &private,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (c *Client) DeleteRepo(name string) error {
	u, _, err := c.Users.Get(context.Background(), c.org)
	if err != nil {
		return err
	}
	_, err = c.Repositories.Delete(context.Background(), *u.Login, name)
	return err
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
		return fmt.Errorf("Actions.CreateOrUpdateRepoSecret returned error: %v", err)
	}

	return nil
}

func (c *Client) AddAWSSecrets(repo string, awsClient *aws.AWS) error {
	awsCredentials, err := awsClient.Credentials()
	if err != nil {
		return err
	}
	if err := c.AddSecret(repo, "AWS_ACCESS_KEY_ID", awsCredentials.AccessKeyID); err != nil {
		return err
	}
	if err := c.AddSecret(repo, "AWS_SECRET_ACCESS_KEY", awsCredentials.SecretAccessKey); err != nil {
		return err
	}
	return nil
}

func replaceImportPaths(repoDir, old, new string) error {
	old = strings.ReplaceAll(old, "https://", "")
	new = strings.ReplaceAll(new, "https://", "")
	return filepath.Walk(repoDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		n := info.Name()
		if strings.HasSuffix(n, ".go") || strings.HasSuffix(n, ".mod") {
			fbuf, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			new := strings.ReplaceAll(string(fbuf), old, new)
			err = ioutil.WriteFile(path, []byte(new), 0)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func addGithubWorkflow(projectPath string) error {
	destFolder := fmt.Sprintf("%s/.github/workflows", projectPath)
	err := os.MkdirAll(destFolder, os.ModePerm)
	if err != nil {
		return err
	}
	workflow, err := assets.Asset("github/mantil-workflow.yml")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/mantil-workflow.yml", destFolder), workflow, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateRepoFromTemplate(
	templateRepo string,
	repoName string,
	path string,
	localConfig *mantil.LocalProjectConfig,
) (string, error) {
	ghRepo, err := c.CreateRepo(repoName, c.org, true)
	if err != nil {
		return "", err
	}
	_, err = git.PlainClone(path, false, &git.CloneOptions{
		URL:      templateRepo,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		return "", err
	}
	err = os.RemoveAll(fmt.Sprintf("%s/.git", path))
	if err != nil {
		return "", err
	}
	repo, err := git.PlainInit(path, false)
	if err != nil {
		return "", err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return "", err
	}
	err = replaceImportPaths(path, templateRepo, *ghRepo.HTMLURL)
	if err != nil {
		return "", err
	}
	err = addGithubWorkflow(repoName)
	if err != nil {
		return "", err
	}
	if err = localConfig.Save(path); err != nil {
		return "", err
	}
	err = wt.AddGlob(".")
	if err != nil {
		return "", err
	}
	_, err = wt.Commit("initial commit", &git.CommitOptions{})
	if err != nil {
		return "", err
	}
	var auth transport.AuthMethod
	var remoteURL string
	sshPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	if _, err := os.Stat(sshPath); err == nil {
		sshKey, _ := ioutil.ReadFile(sshPath)
		auth, err = ssh.NewPublicKeys("git", []byte(sshKey), "")
		if err != nil {
			return "", err
		}
		remoteURL = *ghRepo.SSHURL
	} else {
		auth = &http.BasicAuth{
			Username: "mantil",
			Password: c.token,
		}
		remoteURL = *ghRepo.HTMLURL
	}
	remoteName := "origin"
	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{remoteURL},
	})
	if err != nil {
		return "", err
	}
	err = remote.Push(&git.PushOptions{
		RemoteName: remoteName,
		Auth:       auth,
	})
	if err != nil {
		return "", err
	}
	return *ghRepo.HTMLURL, nil
}
