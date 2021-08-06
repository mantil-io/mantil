package data

import (
	"context"
	"fmt"

	"github.com/atoz-technology/mantil-backend/internal/mantil"
	"github.com/atoz-technology/mantil-backend/internal/stream"
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
	p, err := f.streamingLogsProject(ctx, req.ProjectName)
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

func (f *Data) streamingLogsProject(ctx context.Context, name string) (*mantil.Project, error) {
	var p *mantil.Project
	err := stream.APIGatewayLambdaLogStream(ctx, func() error {
		var err error
		p, err = mantil.LoadProject(name)
		return err
	})
	return p, err
}

func New() *Data {
	return &Data{}
}
