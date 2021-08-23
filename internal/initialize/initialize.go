package initialize

import (
	"fmt"

	"github.com/mantil-io/mantil-backend/internal/aws"
	"github.com/mantil-io/mantil-backend/internal/mantil"
	"github.com/mantil-io/mantil-backend/internal/util"
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
	projectID := mantil.ProjectIdentifier(projectName)
	projectExists, err := doesProjectExist(aws, projectID)
	if err != nil {
		return "", err
	}
	if projectExists {
		return "", fmt.Errorf("project %s already exists", projectName)
	}
	token, err := createProject(aws, projectID, projectName)
	if err != nil {
		return "", err
	}
	return token, nil
}

func doesProjectExist(aws *aws.AWS, name string) (bool, error) {
	return aws.S3BucketExists(name)
}

func createProject(aws *aws.AWS, projectID, name string) (string, error) {
	if err := aws.CreateS3Bucket(projectID, DefaultAWSRegion); err != nil {
		return "", err
	}
	token := util.GenerateToken(TokenLength)
	if token == "" {
		return "", fmt.Errorf("could not generate token for project %s", name)
	}
	if err := saveProjectConfig(name, token); err != nil {
		return "", err
	}
	return token, nil
}

func saveProjectConfig(name, token string) error {
	project, err := mantil.NewProject(name, token)
	if err != nil {
		return fmt.Errorf("could not create project %s config - %v", name, err)
	}
	if err := mantil.SaveProject(project); err != nil {
		return fmt.Errorf("could not save project configuration - %v", err)
	}
	return nil
}
