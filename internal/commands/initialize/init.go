package initialize

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
)

type InitCmd struct {
	aws       *aws.AWS
	name      string
	githubOrg string
}

func New(name, githubOrg string) (*InitCmd, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, fmt.Errorf("could not initialize aws - %v", err)
	}

	return &InitCmd{
		aws:       awsClient,
		name:      name,
		githubOrg: githubOrg,
	}, nil
}

func (i *InitCmd) InitProject() error {
	log.Println("Creating bucket...")
	aws, err := aws.New()
	if err != nil {
		return fmt.Errorf("could not initialize aws - %v", err)
	}
	bucket := mantil.ProjectBucket(i.name)
	bucketExists, err := aws.S3BucketExists(bucket)
	if err != nil {
		return fmt.Errorf("could not check if bucket %s exists - %v", bucket, err)
	}
	if bucketExists {
		return fmt.Errorf("bucket %s already exists", bucket)
	}
	tag := mantil.AccessTag(i.name)
	err = aws.CreateS3Bucket(bucket, "eu-central-1", tag)
	if err != nil {
		return fmt.Errorf("could not create bucket %s - %v", bucket, err)
	}
	log.Println("Creating repo from template...")
	githubClient, err := github.NewClient(i.githubOrg)
	if err != nil {
		return fmt.Errorf("could not initialize github client - %v", err)
	}
	templateRepo := "https://github.com/atoz-technology/go-mantil-template"
	project, err := mantil.NewProject(i.name)
	if err != nil {
		return fmt.Errorf("could not create project %s - %v", i.name, err)
	}
	lc := project.LocalConfig(i.githubOrg)
	repoURL, err := githubClient.CreateRepoFromTemplate(templateRepo, i.name, i.name, lc)
	if err != nil {
		return fmt.Errorf("could not create repo %s from template - %v", i.name, err)
	}
	if err := githubClient.AddAWSSecrets(i.name, aws); err != nil {
		return fmt.Errorf("could not add AWS secrets to repo - %v", err)
	}
	if err := mantil.SaveProject(project); err != nil {
		return fmt.Errorf("could not save project configuration - %v", err)
	}
	projectPath, _ := filepath.Abs(i.name)
	log.Printf("Done!\nProject initialized at %s\nGithub repo URL: %s", projectPath, repoURL)
	return nil
}
