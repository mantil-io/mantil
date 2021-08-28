package data

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/internal/mantil"
)

type Data struct{}

type DataRequest struct {
	ProjectName string
	Token       string
}

type DataResponse struct {
	Project *mantil.Project
}

func (f *Data) Invoke(ctx context.Context, req *DataRequest) (*DataResponse, error) {
	return f.Project(ctx, req)
}

func (f *Data) Project(ctx context.Context, req *DataRequest) (*DataResponse, error) {
	if !f.isRequestValid(req) {
		return nil, fmt.Errorf("bad request")
	}
	p, err := mantil.LoadProject(req.ProjectName)
	if err != nil {
		return nil, err
	}
	if !p.IsValidToken(req.Token) {
		return nil, fmt.Errorf("access denied")
	}
	return &DataResponse{
		Project: p,
	}, nil
}

func (f *Data) isRequestValid(req *DataRequest) bool {
	if req == nil {
		return false
	}
	return req.ProjectName != "" && req.Token != ""
}

func New() *Data {
	return &Data{}
}
