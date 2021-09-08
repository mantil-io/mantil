package git

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/google/go-github/v37/github"
	"github.com/mantil-io/mantil/internal/mantil"
)

func CreateRepoFromTemplate(
	templateRepo, repoName, path string, gc *GithubClient,
	localConfig *mantil.LocalProjectConfig,
) (string, error) {
	if err := cloneTemplate(templateRepo, path); err != nil {
		return "", err
	}
	var gitRepo *git.Repository
	var ghRepo *github.Repository
	var err error
	if gc != nil {
		gitRepo, err = git.PlainInit(path, false)
		if err != nil {
			return "", err
		}
		ghRepo, err = gc.createGithubRepo(repoName, gc.org, true)
		if err != nil {
			return "", err
		}
	}
	newPath := repoName
	if ghRepo != nil {
		newPath = *ghRepo.HTMLURL
	}
	err = replaceImportPaths(path, templateRepo, newPath)
	if err != nil {
		return "", err
	}
	if gc != nil {
		err = gc.addGithubWorkflow(repoName)
		if err != nil {
			return "", err
		}
	}
	if err = localConfig.Save(path); err != nil {
		return "", err
	}
	if gc != nil {
		if err := initRepoCommit(gitRepo); err != nil {
			return "", err
		}
		if err := gc.createRepoRemote(ghRepo, gitRepo); err != nil {
			return "", err
		}
	}
	var repoURL string
	if gc != nil {
		repoURL = *ghRepo.HTMLURL
	}
	return repoURL, nil
}

func cloneTemplate(templateRepo, path string) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      templateRepo,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		return err
	}
	err = os.RemoveAll(fmt.Sprintf("%s/.git", path))
	if err != nil {
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

func initRepoCommit(repo *git.Repository) error {
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = wt.AddGlob(".")
	if err != nil {
		return err
	}
	_, err = wt.Commit("initial commit", &git.CommitOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *GithubClient) createRepoRemote(ghRepo *github.Repository, repo *git.Repository) error {
	auth, remoteURL, err := c.remoteRepoAuth(ghRepo)
	if err != nil {
		return err
	}
	remoteName := "origin"
	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{remoteURL},
	})
	if err != nil {
		return err
	}
	err = remote.Push(&git.PushOptions{
		RemoteName: remoteName,
		Auth:       auth,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *GithubClient) remoteRepoAuth(ghRepo *github.Repository) (transport.AuthMethod, string, error) {
	var auth transport.AuthMethod
	var remoteURL string
	sshPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	if _, err := os.Stat(sshPath); err == nil {
		sshKey, _ := ioutil.ReadFile(sshPath)
		auth, err = ssh.NewPublicKeys("git", []byte(sshKey), "")
		if err != nil {
			return nil, "", err
		}
		remoteURL = *ghRepo.SSHURL
	} else {
		auth = &http.BasicAuth{
			Username: "mantil",
			Password: c.token,
		}
		remoteURL = *ghRepo.HTMLURL
	}
	return auth, remoteURL, nil
}
