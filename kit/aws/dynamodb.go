package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

const (
	dynamodbTableResourceType = "dynamodb:table"
)

func (a *AWS) DeleteDynamodbTablesByTags(tags []TagFilter) error {
	tableARNs, err := a.GetResourcesByTypeAndTag([]string{dynamodbTableResourceType}, tags)
	if err != nil {
		return err
	}
	for _, arn := range tableARNs {
		name, err := dynamodbTableNameFromARN(arn)
		if err != nil {
			return err
		}
		dti := &dynamodb.DeleteTableInput{
			TableName: aws.String(name),
		}
		_, err = a.dynamodbClient.DeleteTable(context.Background(), dti)
		if err != nil {
			return err
		}
	}
	return nil
}

func dynamodbTableNameFromARN(arn string) (string, error) {
	resource, err := resourceFromARN(arn)
	if err != nil {
		return "", err
	}
	// table/{name}
	return strings.TrimPrefix(resource, "table/"), nil
}
