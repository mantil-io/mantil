package initialize

import (
	"fmt"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/config"
)

func InitProject(name string) (*config.Project, error) {
	aws, err := aws.New()
	if err != nil {
		return nil, err
	}
	projectExists, err := config.ProjectExists(name, aws)
	if err != nil {
		return nil, err
	}
	if projectExists {
		return nil, fmt.Errorf("project %s already exists", name)
	}
	project, err := config.CreateProject(name, aws)
	if err != nil {
		return nil, err
	}
	return project, nil
}
