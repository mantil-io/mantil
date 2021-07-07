package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"

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

func githubToken() (string, error) {
	t, ok := os.LookupEnv("GITHUB_TOKEN")
	if ok {
		return t, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
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
		templateUrl := "https://github.com/atoz-technology/go-mantil-template.git"
		_, err = git.PlainClone(project.Name, false, &git.CloneOptions{
			URL:      templateUrl,
			Progress: os.Stdout,
			Depth:    1,
		})
		if err != nil {
			log.Fatal(err)
		}
		err = os.RemoveAll(fmt.Sprintf("%s/.git", project.Name))
		if err != nil {
			log.Fatal(err)
		}
		repo, err := git.PlainInit(project.Name, false)
		if err != nil {
			log.Fatal(err)
		}
		wt, err := repo.Worktree()
		if err != nil {
			log.Fatal(err)
		}
		err = wt.AddGlob(".")
		if err != nil {
			log.Fatal(err)
		}
		_, err = wt.Commit("initial commit", &git.CommitOptions{})
		if err != nil {
			log.Fatal(err)
		}
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
			Name:    &project.Name,
			Private: &private,
		})
		if err != nil {
			log.Fatal(err)
		}
		remoteName := "origin"
		remote, err := repo.CreateRemote(&config.RemoteConfig{
			Name: remoteName,
			URLs: []string{*githubRepo.HTMLURL},
		})
		if err != nil {
			log.Fatal(err)
		}
		err = remote.Push(&git.PushOptions{
			RemoteName: remoteName,
			Auth: &http.BasicAuth{
				Username: "mantil",
				Password: githubAuthToken,
			},
		})
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
