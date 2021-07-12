package cmd

import (
	"log"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/pkg/mantil"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a mantil project",
	Run: func(cmd *cobra.Command, args []string) {
		namePrompt := promptui.Prompt{
			Label: "Project name",
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
		githubClient, err := github.NewClient()
		if err != nil {
			log.Fatal(err)
		}
		templateRepo := "https://github.com/atoz-technology/go-mantil-template"
		if err := githubClient.CreateRepoFromTemplate(templateRepo, projectName); err != nil {
			log.Fatal(err)
		}
		if err := githubClient.AddAWSSecrets(projectName, aws); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
