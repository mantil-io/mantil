package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

func (a *AWS) CreateCloudformationStack(name, templateBody string) error {
	csi := &cloudformation.CreateStackInput{
		StackName:    aws.String(name),
		Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
		OnFailure:    types.OnFailureDelete,
		TemplateBody: aws.String(templateBody),
	}
	cso, err := a.cloudformationClient.CreateStack(context.Background(), csi)
	if err != nil {
		return fmt.Errorf("could not create stack %s - %w", name, err)
	}
	w := cloudformation.NewStackCreateCompleteWaiter(a.cloudformationClient, func(opts *cloudformation.StackCreateCompleteWaiterOptions) {
		opts.MinDelay = 10 * time.Second
		opts.MaxDelay = 20 * time.Second
	})
	dsi := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}
	if err := w.Wait(context.Background(), dsi, 5*time.Minute); err != nil {
		reason, err := a.cloudformationStackActionFailedReason(aws.ToString(cso.StackId))
		if err == nil && reason != "" {
			return fmt.Errorf("could not create stack %s - %s", name, reason)
		}
		return fmt.Errorf("could not create stack %s", name)
	}
	return nil
}

func (a *AWS) DeleteCloudformationStack(name string) error {
	dsi := &cloudformation.DeleteStackInput{
		StackName: aws.String(name),
	}
	_, err := a.cloudformationClient.DeleteStack(context.Background(), dsi)
	if err != nil {
		return fmt.Errorf("could not delete stack %s - %w", name, err)
	}
	w := cloudformation.NewStackDeleteCompleteWaiter(a.cloudformationClient, func(opts *cloudformation.StackDeleteCompleteWaiterOptions) {
		opts.MinDelay = 10 * time.Second
		opts.MaxDelay = 20 * time.Second
	})
	descsi := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}
	if err := w.Wait(context.Background(), descsi, 5*time.Minute); err != nil {
		reason, err := a.cloudformationStackActionFailedReason(name)
		if err == nil {
			return fmt.Errorf("could not delete stack %s - %s", name, reason)
		}
		return fmt.Errorf("could not delete stack %s", name)
	}
	return nil
}

// tries to find reason why stack action failed by going through all the events until first one with status failed is encountered
// asumption is made that first failed event is also the reason for the failure of the stack action
// for stacks that do not longer exists (like in the case of create failure) stack id must be provided instead of name
func (a *AWS) cloudformationStackActionFailedReason(stack string) (string, error) {
	dsei := &cloudformation.DescribeStackEventsInput{
		StackName: aws.String(stack),
	}
	dseo, err := a.cloudformationClient.DescribeStackEvents(context.Background(), dsei)
	if err != nil {
		return "", err
	}
	if len(dseo.StackEvents) == 0 {
		return "", fmt.Errorf("could not find reason why stack %s failed", stack)
	}
	for i := len(dseo.StackEvents) - 1; i >= 0; i-- {
		event := dseo.StackEvents[i]
		if strings.Contains(string(event.ResourceStatus), "FAILED") {
			return fmt.Sprintf("%s - %s", aws.ToString(event.LogicalResourceId), aws.ToString(event.ResourceStatusReason)), nil
		}
	}
	return "", fmt.Errorf("could not find reason why stack %s failed", stack)
}
