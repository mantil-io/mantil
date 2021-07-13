package initialize

import (
	"fmt"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
)

type InitCmd struct {
	name string
}

func New(name string) *InitCmd {
	return &InitCmd{
		name: name,
	}
}

func (i *InitCmd) InitProject() error {
	project, err := mantil.NewProject(i.name, "")
	if err != nil {
		return fmt.Errorf("could not create project %s - %v", i.name, err)
	}
	aws, err := aws.New()
	if err != nil {
		return fmt.Errorf("could not initialize aws - %v", err)
	}
	bucketExists, err := aws.S3BucketExists(project.Bucket)
	if err != nil {
		return fmt.Errorf("could not check if bucket %s exists - %v", project.Bucket, err)
	}
	if bucketExists {
		return fmt.Errorf("bucket %s already exists", project.Bucket)
	}
	err = aws.CreateS3Bucket(project.Bucket, "eu-central-1")
	if err != nil {
		return fmt.Errorf("could not create bucket %s - %v", project.Bucket, err)
	}
	githubClient, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("could not initialize github client - %v", err)
	}
	templateRepo := "https://github.com/atoz-technology/go-mantil-template"
	if err := githubClient.CreateRepoFromTemplate(templateRepo, i.name); err != nil {
		return fmt.Errorf("could not create repo %s from template - %v", i.name, err)
	}
	if err := githubClient.AddAWSSecrets(i.name, aws); err != nil {
		return fmt.Errorf("could not add AWS secrets to repo - %v", err)
	}
	return nil
}
