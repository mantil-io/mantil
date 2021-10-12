package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/mantil-io/mantil/cli/log"
)

type CloudFormation struct {
	aws *AWS
	cli *cloudformation.Client
}

func (a *AWS) CloudFormation() *CloudFormation {
	return &CloudFormation{
		aws: a,
		cli: a.cloudformationClient,
	}
}

func (f *CloudFormation) CreateStack(name, templateBody string, resourceTags map[string]string) error {
	csi := &cloudformation.CreateStackInput{
		StackName:    aws.String(name),
		Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
		OnFailure:    types.OnFailureDelete,
		TemplateBody: aws.String(templateBody),
	}

	if resourceTags != nil {
		tags := []types.Tag{}
		for k, v := range resourceTags {
			tags = append(tags, types.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
		csi.Tags = tags
	}

	cso, err := f.cli.CreateStack(context.Background(), csi)
	if err != nil {
		return log.Wrap(err, fmt.Sprintf("could not create stack %s", name))
	}
	w := cloudformation.NewStackCreateCompleteWaiter(f.cli, func(opts *cloudformation.StackCreateCompleteWaiterOptions) {
		opts.MinDelay = 10 * time.Second
		opts.MaxDelay = 20 * time.Second
	})
	dsi := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}
	if err := w.Wait(context.Background(), dsi, 5*time.Minute); err != nil {
		reason := f.stackActionFailedReason(aws.ToString(cso.StackId))
		if reason != "" {
			return log.Wrap(fmt.Errorf("could not create stack %s - %s", name, reason))
		}
		return log.Wrap(fmt.Errorf("could not create stack %s", name))
	}
	return nil
}

func (f *CloudFormation) DeleteStack(name string) error {
	dsi := &cloudformation.DeleteStackInput{
		StackName: aws.String(name),
	}
	_, err := f.cli.DeleteStack(context.Background(), dsi)
	if err != nil {
		return log.Wrap(err, fmt.Sprintf("could not delete stack %s", name))
	}
	w := cloudformation.NewStackDeleteCompleteWaiter(f.cli, func(opts *cloudformation.StackDeleteCompleteWaiterOptions) {
		opts.MinDelay = 10 * time.Second
		opts.MaxDelay = 20 * time.Second
	})
	descsi := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}
	if err := w.Wait(context.Background(), descsi, 5*time.Minute); err != nil {
		reason := f.stackActionFailedReason(name)
		if reason != "" {
			return log.Wrap(fmt.Errorf("could not delete stack %s - %s", name, reason))
		}
		return log.Wrap(fmt.Errorf("could not delete stack %s", name))
	}
	return nil
}

// tries to find reason why stack action failed by going through all the events until first one with status failed is encountered
// asumption is made that first failed event is also the reason for the failure of the stack action
// for stacks that do not longer exists (like in the case of create failure) stack id must be provided instead of name
func (f *CloudFormation) stackActionFailedReason(stack string) string {
	dsei := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stack),
	}
	dseo, err := f.cli.DescribeStackEvents(context.Background(), dsei)
	if err != nil {
		log.Printf("DescribeStackEvents error: %v", err)
		return ""
	}
	if len(dseo.StackEvents) == 0 {
		return ""
	}
	for i := len(dseo.StackEvents) - 1; i >= 0; i-- {
		event := dseo.StackEvents[i]
		if strings.Contains(string(event.ResourceStatus), "FAILED") {
			return fmt.Sprintf("%s - %s",
				aws.ToString(event.LogicalResourceId),
				aws.ToString(event.ResourceStatusReason))
		}
	}
	return ""
}
