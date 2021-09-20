package initialize

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/mantil"
)

func InitProject(name string) (string, error) {
	aws, err := aws.New()
	if err != nil {
		return "", err
	}
	projectExists, err := mantil.ProjectExists(name, aws)
	if err != nil {
		return "", err
	}
	if projectExists {
		return "", fmt.Errorf("project %s already exists", name)
	}
	project, err := mantil.CreateProject(name, aws)
	if err != nil {
		return "", err
	}
	return project.Token, nil
}
