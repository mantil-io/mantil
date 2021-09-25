package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/mantil-io/mantil/auth"
)

func generatePolicy(principalId, effect, resource string) *events.APIGatewayCustomAuthorizerResponse {
	rsp := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}

	if effect != "" && resource != "" {
		rsp.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}
	return &rsp
}

func allow(req *events.APIGatewayCustomAuthorizerRequestTypeRequest) *events.APIGatewayCustomAuthorizerResponse {
	return generatePolicy("Mantil", "Allow", req.MethodArn)
}

func errorResponse(err error) (*events.APIGatewayCustomAuthorizerResponse, error) {
	log.Print(err)
	return nil, err
}

func handleRequest(ctx context.Context, req *events.APIGatewayCustomAuthorizerRequestTypeRequest) (*events.APIGatewayCustomAuthorizerResponse, error) {
	buf, _ := json.Marshal(req)
	log.Printf("req %s", buf)

	claims, err := auth.ReadAccessToken(req)
	if err != nil {
		return errorResponse(fmt.Errorf("read runtime access token error %w", err))
	}
	rsp := allow(req)
	auth.StoreUserClaims(claims, rsp)
	return rsp, nil
}

func main() {
	lambda.Start(handleRequest)
}
