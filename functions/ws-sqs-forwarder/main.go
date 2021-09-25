package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mantil-io/mantil/backend/api/ws"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event events.SQSEvent) error {
	h, err := ws.NewHandler()
	if err != nil {
		return err
	}
	return h.HandleSQSEvent(ctx, event)
}
