package aws

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
)

func (a *AWS) PublishToAPIGatewayConnection(domain, stage, connectionID string, data []byte) error {
	cfg := a.config.Copy()
	cfg.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if service != "execute-api" {
			return cfg.EndpointResolver.ResolveEndpoint(service, region)
		}
		var endpoint url.URL
		endpoint.Path = stage
		endpoint.Host = domain
		endpoint.Scheme = "https"
		return aws.Endpoint{
			SigningRegion: region,
			URL:           endpoint.String(),
		}, nil
	})
	agwClient := apigatewaymanagementapi.NewFromConfig(cfg)
	ptci := &apigatewaymanagementapi.PostToConnectionInput{
		ConnectionId: aws.String(connectionID),
		Data:         data,
	}
	_, err := agwClient.PostToConnection(context.Background(), ptci)
	return err
}
