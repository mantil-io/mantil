package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/api/security"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	api := security.New()

	dec, err := base64.StdEncoding.DecodeString(event.QueryStringParameters[dto.RequestQueryParam])
	if err != nil {
		return errorResponse(err), nil
	}
	var req dto.SecurityRequest
	if err := json.Unmarshal(dec, &req); err != nil {
		return errorResponse(err), nil
	}

	resp, err := api.Invoke(context.Background(), req)
	if err != nil {
		return errorResponse(err), nil
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return errorResponse(err), nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(b),
	}, nil
}

func errorResponse(err error) events.APIGatewayProxyResponse {
	//  https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/credentials/endpointcreds
	rsp := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{
		Code:    "InternalServerError",
		Message: err.Error(),
	}
	buf, _ := json.Marshal(rsp)
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusInternalServerError,
		Body:       string(buf),
	}
}
