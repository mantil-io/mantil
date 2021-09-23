package initialize

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/mantil"
)

func InitProject(name string) (*mantil.Project, error) {
	aws, err := aws.New()
	if err != nil {
		return nil, err
	}
	projectExists, err := mantil.ProjectExists(name, aws)
	if err != nil {
		return nil, err
	}
	if projectExists {
		return nil, fmt.Errorf("project %s already exists", name)
	}
	project, err := mantil.CreateProject(name, aws)
	if err != nil {
		return nil, err
	}
	return project, nil
}
