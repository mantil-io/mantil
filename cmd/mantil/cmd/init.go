package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atoz-technology/mantil-cli/pkg/mantil"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/github"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

const templateRepo = "github.com/atoz-technology/go-mantil-template"

func githubToken() (string, error) {
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

func replaceImportPaths(projectDir string, repoURL string) error {
	repoURL = strings.ReplaceAll(repoURL, "https://", "")
	return filepath.Walk(projectDir, func(path string, info fs.FileInfo, err error) error {
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
			new := strings.ReplaceAll(string(fbuf), templateRepo, repoURL)
			err = ioutil.WriteFile(path, []byte(new), 0)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func createRepoFromTemplate(projectName string) error {
	githubAuthToken, err := githubToken()
	if err != nil {
		log.Fatal("Could not find GitHub access token")
	}
	githubClient := github.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubAuthToken}),
		),
	)
	private := true
	githubRepo, _, err := githubClient.Repositories.Create(context.Background(), "", &github.Repository{
		Name:    &projectName,
		Private: &private,
	})
	if err != nil {
		return err
	}
	templateUrl := fmt.Sprintf("https://%s.git", templateRepo)
	_, err = git.PlainClone(projectName, false, &git.CloneOptions{
		URL:      templateUrl,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		return err
	}
	err = os.RemoveAll(fmt.Sprintf("%s/.git", projectName))
	if err != nil {
		return err
	}
	repo, err := git.PlainInit(projectName, false)
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	err = replaceImportPaths(projectName, *githubRepo.HTMLURL)
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
	remoteName := "origin"
	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{*githubRepo.HTMLURL},
	})
	if err != nil {
		return err
	}
	err = remote.Push(&git.PushOptions{
		RemoteName: remoteName,
		Auth: &http.BasicAuth{
			Username: "mantil",
			Password: githubAuthToken,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a mantil project",
	Run: func(cmd *cobra.Command, args []string) {
		namePrompt := promptui.Prompt{
			Label:   "Project name",
			Default: "test",
		}
		projectName, err := namePrompt.Run()
		if err != nil {
			log.Fatal(err)
		}
		project := mantil.NewProject(projectName)
		// TODO check project name availability
		// cfg, err := config.LoadDefaultConfig(context.TODO())
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// client := s3.NewFromConfig(cfg)
		// _, err = client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		// 	Bucket: aws.String(project.Bucket),
		// })
		// if err != nil {
		// 	log.Fatal(err)
		// }
		err = createRepoFromTemplate(project.Name)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
