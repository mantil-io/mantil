package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/aws"
)

func handleRequest(ctx context.Context, req *events.APIGatewayCustomAuthorizerRequestTypeRequest) (*events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	buf, _ := json.Marshal(req)
	log.Printf("req %s", buf)

	pk, err := publicKey()
	if err != nil {
		return errorResponse(err)
	}
	claims, err := domain.ReadAccessToken(req.Headers, pk)
	if err != nil {
		return errorResponse(fmt.Errorf("read runtime access token error %w", err))
	}
	rsp := allowResponse(claims)
	buf, _ = json.Marshal(rsp)
	log.Printf("rsp %s", buf)
	return rsp, nil
}

func publicKey() (string, error) {
	pk := os.Getenv(domain.EnvPublicKey)
	if pk != "" {
		return pk, nil
	}
	awsClient, err := aws.New()
	if err != nil {
		return "", err
	}
	path, err := domain.SSMParameterPath(domain.SSMPublicKey)
	if err != nil {
		return "", err
	}
	pk, err = awsClient.GetSSMParameter(path)
	if err != nil {
		return "", err
	}
	return pk, nil
}

func allowResponse(claims *domain.AccessTokenClaims) *events.APIGatewayV2CustomAuthorizerSimpleResponse {
	rsp := &events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context:      make(map[string]interface{}),
	}
	domain.StoreUserClaims(claims, rsp.Context)
	return rsp
}

func errorResponse(err error) (*events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	log.Print(err)
	return nil, err
}

func main() {
	lambda.Start(handleRequest)
}
