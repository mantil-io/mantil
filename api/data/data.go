package data

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/config"
)

type Data struct {
	project *config.Project
}

func New() *Data {
	return &Data{}
}

func (d *Data) Invoke(ctx context.Context, req *dto.DataRequest) (*dto.DataResponse, error) {
	if err := d.init(req); err != nil {
		return nil, err
	}
	return d.data()
}

func (d *Data) init(req *dto.DataRequest) error {
	project, err := config.LoadProjectS3(req.ProjectName)
	if err != nil {
		return fmt.Errorf("error fetching project %s - %w", req.ProjectName, err)
	}
	d.project = project
	return nil
}

func (d *Data) data() (*dto.DataResponse, error) {
	return &dto.DataResponse{
		Project: d.project,
	}, nil
}
