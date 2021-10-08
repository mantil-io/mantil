package data

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/workspace"
)

type Data struct {
	stage *workspace.Stage
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
	stage, err := workspace.LoadStageState(req.ProjectName, req.StageName)
	if err != nil {
		return fmt.Errorf("error fetching stage %s for project %s - %w", req.StageName, req.ProjectName, err)
	}
	d.stage = stage
	return nil
}

func (d *Data) data() (*dto.DataResponse, error) {
	return &dto.DataResponse{
		Stage: d.stage,
	}, nil
}
