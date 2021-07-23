package initialize

import (
	"fmt"

	"github.com/atoz-technology/mantil-backend/internal/aws"
	"github.com/atoz-technology/mantil-backend/internal/mantil"
	"github.com/atoz-technology/mantil-backend/internal/util"
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

	projectId := mantil.ProjectIdentifier(projectName)
	projectExists, err := aws.S3BucketExists(projectId)
	if err != nil {
		return "", err
	}
	if projectExists {
		return "", fmt.Errorf("project %s already exists", projectId)
	}

	err = aws.CreateS3Bucket(projectId, DefaultAWSRegion)
	if err != nil {
		return "", err
	}

	token := util.GenerateToken(TokenLength)
	if token == "" {
		return "", fmt.Errorf("could not generate token for project %s - %v", projectId, err)
	}

	project, err := mantil.NewProject(projectName, token)
	if err != nil {
		return "", fmt.Errorf("could not create project %s - %v", projectId, err)
	}

	if err := mantil.SaveProject(project); err != nil {
		return "", fmt.Errorf("could not save project configuration - %v", err)
	}

	return token, nil
}
