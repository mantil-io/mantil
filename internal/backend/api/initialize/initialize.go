package initialize

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/mantil"
	"github.com/mantil-io/mantil/internal/util"
)

const (
	TokenLength         = 40
	DefaultAWSRegion    = "eu-central-1"
	ProjectPartitionKey = "project"
)

func InitProject(projectName string) (string, error) {
	aws, err := aws.New()
	if err != nil {
		return "", err
	}
	bucket, err := mantil.ProjectBucket(projectName, aws)
	if err != nil {
		return "", err
	}
	projectExists, err := doesProjectExist(aws, bucket)
	if err != nil {
		return "", err
	}
	if projectExists {
		return "", fmt.Errorf("project %s already exists", projectName)
	}
	token, err := createProject(aws, bucket, projectName)
	if err != nil {
		return "", err
	}
	return token, nil
}

func doesProjectExist(aws *aws.AWS, name string) (bool, error) {
	return aws.S3BucketExists(name)
}

func createProject(aws *aws.AWS, bucket, name string) (string, error) {
	if err := aws.CreateS3Bucket(bucket, DefaultAWSRegion); err != nil {
		return "", err
	}
	token := util.GenerateToken(TokenLength)
	if token == "" {
		return "", fmt.Errorf("could not generate token for project %s", name)
	}
	project, err := mantil.NewProject(name, token, aws)
	if err != nil {
		return "", err
	}
	if err := saveProjectConfig(project); err != nil {
		return "", err
	}
	return token, nil
}

func saveProjectConfig(project *mantil.Project) error {
	if err := mantil.SaveProject(project); err != nil {
		return fmt.Errorf("could not save project configuration - %v", err)
	}
	return nil
}
