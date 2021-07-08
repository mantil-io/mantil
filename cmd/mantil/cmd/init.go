package cmd

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/atoz-technology/mantil-cli/internal/assets"
	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/pkg/mantil"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const templateRepo = "github.com/atoz-technology/go-mantil-template"

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

func createRepoFromTemplate(projectName string) error {
	githubClient, err := github.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	repoURL, err := githubClient.CreateRepo(projectName, "", true)
	if err != nil {
		log.Fatal(err)
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
	err = replaceImportPaths(projectName, repoURL)
	if err != nil {
		return err
	}
	err = addGithubWorkflow(projectName)
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
		URLs: []string{repoURL},
	})
	if err != nil {
		return err
	}
	err = remote.Push(&git.PushOptions{
		RemoteName: remoteName,
		Auth: &http.BasicAuth{
			Username: "mantil",
			Password: githubClient.Token,
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
		aws, err := aws.New()
		if err != nil {
			log.Fatal(err)
		}
		bucketExists, err := aws.S3BucketExists(project.Bucket)
		if err != nil {
			log.Fatal(err)
		}
		if bucketExists {
			log.Fatal("Project already exists")
		}
		err = aws.CreateS3Bucket(project.Bucket, "eu-central-1")
		if err != nil {
			log.Fatal(err)
		}
		err = createRepoFromTemplate(project.Name)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
