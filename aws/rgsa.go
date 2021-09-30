package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
)

func (a *AWS) GetResourcesByTypeAndTag(resourceTypes []string, resourceTags []TagFilter) ([]string, error) {
	var resourceARNs []string
	var tagFilters []types.TagFilter
	for _, t := range resourceTags {
		tagFilters = append(tagFilters, types.TagFilter{Key: aws.String(t.Key), Values: t.Values})
	}
	gri := &resourcegroupstaggingapi.GetResourcesInput{
		ResourceTypeFilters: resourceTypes,
		TagFilters:          tagFilters,
	}
	for {
		gro, err := a.rgsaClient.GetResources(context.Background(), gri)
		if err != nil {
			return nil, err
		}
		for _, r := range gro.ResourceTagMappingList {
			resourceARNs = append(resourceARNs, *r.ResourceARN)
		}
		if gro.PaginationToken == nil || aws.ToString(gro.PaginationToken) == "" {
			break
		}
		gri.PaginationToken = gro.PaginationToken
	}
	return resourceARNs, nil
}

type TagFilter struct {
	Key    string
	Values []string
}
