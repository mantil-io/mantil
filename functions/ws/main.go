package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mantil-io/mantil/internal/backend/api/ws"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, req interface{}) (interface{}, error) {
	h, err := ws.NewHandler()
	if err != nil {
		return nil, err
	}
	switch req.(type) {
	case events.APIGatewayProxyRequest:
		return h.HandleApiGatewayRequest(ctx, req.(events.APIGatewayWebsocketProxyRequest))
	case events.SQSEvent:
		return nil, h.HandleSQSEvent(ctx, req.(events.SQSEvent))
	}
	return nil, fmt.Errorf("unknown event type")
}
