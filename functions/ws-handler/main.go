package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mantil-io/mantil/internal/backend/api/ws"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	h, err := ws.NewHandler()
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}
	return h.HandleApiGatewayRequest(ctx, event)
}
