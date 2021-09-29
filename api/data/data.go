package data

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/config"
)

type DataRequest struct {
	ProjectName string
	StageName   string
}

type DataResponse struct {
	Stage *config.Stage
}

type Data struct {
	stage *config.Stage
}

func New() *Data {
	return &Data{}
}

func (d *Data) Invoke(ctx context.Context, req *DataRequest) (*DataResponse, error) {
	if err := d.init(req); err != nil {
		return nil, err
	}
	return d.data()
}

func (d *Data) init(req *DataRequest) error {
	stage, err := config.LoadDeploymentState(req.ProjectName, req.StageName)
	if err != nil {
		return fmt.Errorf("error fetching stage %s for project %s - %w", req.StageName, req.ProjectName, err)
	}
	d.stage = stage
	return nil
}

func (d *Data) data() (*DataResponse, error) {
	return &DataResponse{
		Stage: d.stage,
	}, nil
}
