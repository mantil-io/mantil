package data

import (
	"context"
	"fmt"
	"log"

	"github.com/atoz-technology/mantil-backend/internal/mantil"
	"github.com/atoz-technology/mantil-backend/internal/stream"
)

type Data struct {
}

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
	if req.ProjectName == "" || req.Token == "" {
		return nil, fmt.Errorf("bad request")
	}

	var p *mantil.Project
	err := stream.LambdaLogStream(ctx, func() error {
		var err error
		p, err = mantil.LoadProject(req.ProjectName)
		if err != nil {
			log.Printf("%v", err)
			return err
		}

		if p.Token != req.Token {
			log.Printf("access denied - %s", p.Token)
			return fmt.Errorf("access denied")
		}
		return nil
	})
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}

	return &DataResponse{
		Project: p,
	}, nil
}

func New() *Data {
	return &Data{}
}
