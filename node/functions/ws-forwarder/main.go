package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mantil-io/mantil.go/proto"
	"github.com/mantil-io/mantil/node/api/ws"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, msg proto.Message) error {
	h, err := ws.NewHandler()
	if err != nil {
		return err
	}
	return h.HandleBackendMessage(msg)
}
