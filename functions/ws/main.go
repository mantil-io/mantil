package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mantil-io/mantil/internal/backend/api/ws"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event interface{}) (interface{}, error) {
	h, err := ws.NewHandler()
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	gwr := events.APIGatewayWebsocketProxyRequest{}
	if err := json.Unmarshal(data, &gwr); err == nil {
		return h.HandleApiGatewayRequest(ctx, gwr)
	}
	sqsr := events.SQSEvent{}
	if err := json.Unmarshal(data, &sqsr); err == nil {
		return nil, h.HandleSQSEvent(ctx, sqsr)
	}
	return nil, fmt.Errorf("unknown event type")
}
