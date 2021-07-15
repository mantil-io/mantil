package initialize

import (
	"fmt"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/github"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
)

type InitCmd struct {
	aws  *aws.AWS
	name string
}

func New(name string) (*InitCmd, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, fmt.Errorf("could not initialize aws - %v", err)
	}

	return &InitCmd{
		aws:  awsClient,
		name: name,
	}, nil
}

func (i *InitCmd) InitProject() error {
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
	err = aws.CreateS3Bucket(bucket, "eu-central-1")
	if err != nil {
		return fmt.Errorf("could not create bucket %s - %v", bucket, err)
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
	project, err := mantil.NewProject(i.name)
	if err != nil {
		return fmt.Errorf("could not create project %s - %v", i.name, err)
	}
	if err := mantil.SaveProject(project); err != nil {
		return fmt.Errorf("could not save project configuration - %v", err)
	}
	return nil
}
