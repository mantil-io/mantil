package aws

import (
	"context"
	"fmt"
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
	_, err := a.cloudformationClient.CreateStack(context.Background(), csi)
	if err != nil {
		return fmt.Errorf("error creating stack %s - %w", name, err)
	}
	w := cloudformation.NewStackCreateCompleteWaiter(a.cloudformationClient, func(opts *cloudformation.StackCreateCompleteWaiterOptions) {
		opts.MinDelay = 10 * time.Second
		opts.MaxDelay = 20 * time.Second
	})
	dsi := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}
	if err := w.Wait(context.Background(), dsi, 5*time.Minute); err != nil {
		// TODO if waiter fails find a reason why stack create failed
		return fmt.Errorf("error creating stack %s - %w", name, err)
	}
	return nil
}

func (a *AWS) DeleteCloudformationStack(name string) error {
	dsi := &cloudformation.DeleteStackInput{
		StackName: aws.String(name),
	}
	_, err := a.cloudformationClient.DeleteStack(context.Background(), dsi)
	if err != nil {
		return fmt.Errorf("error deleting stack %s - %w", name, err)
	}
	w := cloudformation.NewStackDeleteCompleteWaiter(a.cloudformationClient, func(opts *cloudformation.StackDeleteCompleteWaiterOptions) {
		opts.MinDelay = 10 * time.Second
		opts.MaxDelay = 20 * time.Second
	})
	descsi := &cloudformation.DescribeStacksInput{
		StackName: aws.String(name),
	}
	if err := w.Wait(context.Background(), descsi, 5*time.Minute); err != nil {
		// TODO if waiter fails find a reason why stack delete failed
		return fmt.Errorf("error waiting for stack %s to delete - %w", name, err)
	}
	return nil
}
