package data

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/internal/config"
)

type Data struct{}

type DataRequest struct {
	ProjectName string
}

type DataResponse struct {
	Project *config.Project
}

func (f *Data) Invoke(ctx context.Context, req *DataRequest) (*DataResponse, error) {
	return f.Project(ctx, req)
}

func (f *Data) Project(ctx context.Context, req *DataRequest) (*DataResponse, error) {
	if !f.isRequestValid(req) {
		return nil, fmt.Errorf("bad request")
	}
	p, err := config.LoadProjectS3(req.ProjectName)
	if err != nil {
		return nil, err
	}
	return &DataResponse{
		Project: p,
	}, nil
}

func (f *Data) isRequestValid(req *DataRequest) bool {
	if req == nil {
		return false
	}
	return req.ProjectName != ""
}

func New() *Data {
	return &Data{}
}
