package data

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/config"
)

type Data struct{}

type DataRequest struct {
	ProjectName string
	StageName   string
}

type DataResponse struct {
	Stage *config.Stage
}

func (f *Data) Invoke(ctx context.Context, req *DataRequest) (*DataResponse, error) {
	return f.Stage(ctx, req)
}

func (f *Data) Stage(ctx context.Context, req *DataRequest) (*DataResponse, error) {
	if !f.isRequestValid(req) {
		return nil, fmt.Errorf("bad request")
	}
	s, err := config.LoadDeploymentState(req.ProjectName, req.StageName)
	if err != nil {
		return nil, err
	}
	return &DataResponse{
		Stage: s,
	}, nil
}

func (f *Data) isRequestValid(req *DataRequest) bool {
	if req == nil {
		return false
	}
	return req.ProjectName != "" && req.StageName != ""
}

func New() *Data {
	return &Data{}
}
